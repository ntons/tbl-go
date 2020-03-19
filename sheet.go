package tbl

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"

	pb "github.com/golang/protobuf/proto"
)

type Sheet struct {
	// message type
	t reflect.Type
	// message data
	data []pb.Message
	//
	indexes map[string]*index
}

func LoadFile(t reflect.Type, tblPath string) (s *Sheet, err error) {
	s = &Sheet{
		t:       t,
		indexes: make(map[string]*index),
	}
	if err = s.load(tblPath); err != nil {
		return nil, err
	}
	return
}

func (s *Sheet) load(tblPath string) (err error) {
	var r *bytes.Reader
	if b, err := ioutil.ReadFile(tblPath); err != nil {
		return err
	} else {
		r = bytes.NewReader(b)
	}
	for {
		var n int32
		if err = binary.Read(r, binary.BigEndian, &n); err != nil {
			if err != io.EOF {
				return
			}
			err = nil
			break
		}
		var b = make([]byte, n)
		if _n, err := r.Read(b); err != nil {
			return err
		} else if _n != int(n) {
			return fmt.Errorf(
				"bad read length, expected %v, read %v", n, _n)
		}
		m := reflect.New(s.t).Interface().(pb.Message)
		if err = pb.Unmarshal(b, m); err != nil {
			return
		}
		s.data = append(s.data, m)
	}
	return
}

func (s *Sheet) CreateIndex(field string) (err error) {
	var idx *index
	for _, m := range s.data {
		v := reflect.ValueOf(m).Elem().FieldByName(field)
		if !v.IsValid() {
			return fmt.Errorf("invalid field %s", field)
		}
		if idx == nil {
			idx = newIndex(v.Type())
		}
		idx.add(v.Interface(), m)
	}
	s.indexes[field] = idx
	return
}

func (s *Sheet) FindByIndex(field string, key interface{}) (pb.Message, error) {
	if idx, ok := s.indexes[field]; !ok {
		return nil, fmt.Errorf("no such index %s", field)
	} else {
		return idx.get(key)
	}
}

func (s *Sheet) Foreach(callback func(m pb.Message) error) (err error) {
	for _, m := range s.data {
		if err = callback(m); err != nil {
			return
		}
	}
	return
}
