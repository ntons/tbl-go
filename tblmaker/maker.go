package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/tealeg/xlsx"
)

const (
	ProtoIndent = "    "
	ProtoExt    = ".proto"
	TableExt    = ".tbl"
)

type Maker struct {
	FilePath        string
	FileName        string
	PackageName     string
	ImportPaths     StringSlice
	ImportFiles     StringSlice
	Sheet           *xlsx.Sheet
	ProtoOutputPath string
	TableOutputPath string
}

func (x *Maker) Make() (err error) {
	if x.FileName == "" {
		x.FileName = TrimExt(filepath.Base(x.FilePath))
	}
	if f, err := os.Stat(x.ProtoOutputPath); err != nil {
		return err
	} else if f.IsDir() {
		x.ProtoOutputPath = filepath.Join(
			x.ProtoOutputPath, Underscore(x.FileName, x.Sheet.Name)+ProtoExt)
	}
	if f, err := os.Stat(x.TableOutputPath); err != nil {
		return err
	} else if f.IsDir() {
		x.TableOutputPath = filepath.Join(
			x.TableOutputPath, Underscore(x.FileName, x.Sheet.Name)+TableExt)
	}
	if err = x.makeProto(); err != nil {
		return
	}
	if err = x.makeTable(); err != nil {
		return
	}
	return
}

func (x *Maker) makeProto() (err error) {
	if x.Sheet.MaxRow < 2 {
		return fmt.Errorf("no field desc row")
	}
	buf := bytes.NewBufferString(fmt.Sprintf(`
// This file was generated automatically by tblmaker.
// DO NOT edit it manually!
// Created At: ` + time.Now().Format("2006-01-02 15:04:05") + `

syntax = "proto3";

package ` + x.PackageName + `;

`))
	for _, file := range x.ImportFiles {
		buf.WriteString(fmt.Sprintf("import %q;\n", file))
	}
	buf.WriteByte('\n')

	fs := make([]string, 0, x.Sheet.MaxCol)
	for i := 0; i < x.Sheet.MaxCol; i++ {
		if f := x.Sheet.Cell(1, i).String(); f != "" {
			fs = append(fs, f)
		}
	}
	x.makeProtoMessage(buf, 0, UpperCamel(x.FileName, x.Sheet.Name), fs)
	if err = ioutil.WriteFile(x.ProtoOutputPath, buf.Bytes(), 0644); err != nil {
		return
	}
	return
}

func (x *Maker) makeProtoMessage(
	buf *bytes.Buffer, dp int, name string, fs []string) (err error) {
	if dp > 1 {
		return fmt.Errorf("nested of nested message is forbiddened")
	}
	var (
		indent0 = strings.Repeat(ProtoIndent, dp)
		indent1 = strings.Repeat(ProtoIndent, dp+1)
	)
	buf.WriteString(fmt.Sprintf("%smessage %s {\n", indent0, name))
	for _, f := range fs {
		var label, type_, name, num string
		if a := strings.SplitN(f, "#", 3); len(a) != 3 {
			return fmt.Errorf("bad field expr: %q", f)
		} else {
			num, name, type_ = a[0], a[1], a[2]
			if _, err = strconv.Atoi(num); err != nil {
				return fmt.Errorf("bad field expr: %q", f)
			}
			for _, c := range name {
				if ('a' <= c && c <= 'z') ||
					('0' <= c && c <= '9') ||
					(c == '_') {
					continue
				}
				return fmt.Errorf("bad field expr: %q", f)
			}
		}
		// []xxx repeated field
		if strings.HasPrefix(type_, "[]") {
			label, type_ = "repeated ", type_[2:]
		}
		// xxx{} nested message
		if strings.HasSuffix(type_, "}") {
			k := strings.Index(type_, "{")
			fs := strings.Split(type_[k+1:len(type_)-1], ";")
			if err = x.makeProtoMessage(buf, dp+1, type_[:k], fs); err != nil {
				return
			}
			type_ = type_[:k]
		}
		buf.WriteString(fmt.Sprintf(
			"%s%s%s %s = %s;\n", indent1, label, type_, name, num))
	}
	buf.WriteString(fmt.Sprintf("%s}\n", indent0))
	return
}

func (x *Maker) makeTable() (err error) {
	p := protoparse.Parser{
		ImportPaths: append(x.ImportPaths, filepath.Dir(x.ProtoOutputPath)),
	}
	fds, err := p.ParseFiles(filepath.Base(x.ProtoOutputPath))
	if err != nil {
		return
	}
	md := fds[0].GetMessageTypes()[0]
	// prepare col num
	colNum := make([]int, x.Sheet.MaxCol)
	for col := 0; col < x.Sheet.MaxCol; col++ {
		if f := x.Sheet.Cell(1, col).String(); f == "" {
			continue
		} else if k := strings.Index(f, "#"); k < 0 {
			continue
		} else if n, err := strconv.Atoi(f[:k]); err != nil {
			continue
		} else {
			colNum[col] = n
		}
	}
	// dump data
	buf := bytes.NewBuffer(nil)
	for row := 2; row < x.Sheet.MaxRow; row++ {
		fs := make(map[int]string)
		for col := 0; col < x.Sheet.MaxCol; col++ {
			if colNum[col] == 0 {
				continue
			}
			fs[colNum[col]] = x.Sheet.Cell(row, col).String()
		}
		m, err := x.makeTableMessage(md, fs)
		if err != nil {
			return err
		}
		b, err := m.Marshal()
		if err != nil {
			return err
		}
		binary.Write(buf, binary.BigEndian, int32(len(b)))
		buf.Write(b)
	}
	if err = ioutil.WriteFile(x.TableOutputPath, buf.Bytes(), 0644); err != nil {
		return
	}
	return
}

func (x *Maker) makeTableMessage(
	md *desc.MessageDescriptor, fs map[int]string) (
	m *dynamic.Message, err error) {
	m = dynamic.NewMessage(md)
	for n, f := range fs {
		fd := md.FindFieldByNumber(int32(n))
		if fd == nil {
			return nil, fmt.Errorf("invalid message number: %d", n)
		}
		if t := fd.GetType(); t == dpb.FieldDescriptorProto_TYPE_MESSAGE {
			_md := fd.GetMessageType()
			if _md == nil {
				return nil, fmt.Errorf("no message type")
			}
			if fd.GetLabel() == dpb.FieldDescriptorProto_LABEL_REPEATED {
				for _, _f := range strings.Split(f, ";") {
					_fs := make(map[int]string)
					for _i, __f := range strings.Split(_f, ",") {
						_fs[_i+1] = __f
					}
					v, err := x.makeTableMessage(_md, _fs)
					if err != nil {
						return nil, err
					}
					m.AddRepeatedFieldByNumber(n, v)
				}
			} else {
				_fs := make(map[int]string)
				for _i, _f := range strings.Split(f, ",") {
					_fs[_i+1] = _f
				}
				v, err := x.makeTableMessage(_md, _fs)
				if err != nil {
					return nil, err
				}
				m.SetFieldByNumber(n, v)
			}
		} else {
			if fd.GetLabel() == dpb.FieldDescriptorProto_LABEL_REPEATED {
				for _, _f := range strings.Split(f, ",") {
					v, err := TypedValue(_f, t)
					if err != nil {
						return nil, err
					}
					m.AddRepeatedFieldByNumber(n, v)
				}
			} else {
				v, err := TypedValue(f, t)
				if err != nil {
					return nil, err
				}
				m.SetFieldByNumber(n, v)
			}
		}
	}
	return
}
