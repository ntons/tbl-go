package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
)

func Underscore(a ...string) (s string) {
	for i, _ := range a {
		if len(s) == 0 {
			s += strings.ToLower(a[i])
		} else {
			s += "_" + strings.ToLower(a[i])
		}
	}
	return strings.ToLower(s)
}

func UpperCamel(a ...string) (s string) {
	for i, _ := range a {
		s += strings.Title(strings.ToLower(a[i]))
	}
	return
}

func TrimExt(s string) string {
	return strings.TrimSuffix(s, filepath.Ext(s))
}

func TypedValue(
	s string, t dpb.FieldDescriptorProto_Type) (interface{}, error) {
	switch t {
	case dpb.FieldDescriptorProto_TYPE_DOUBLE,
		dpb.FieldDescriptorProto_TYPE_FIXED64:
		if v, err := strconv.ParseFloat(s, 64); err != nil {
			return nil, err
		} else {
			return v, nil
		}
	case dpb.FieldDescriptorProto_TYPE_FLOAT,
		dpb.FieldDescriptorProto_TYPE_FIXED32:
		if v, err := strconv.ParseFloat(s, 32); err != nil {
			return nil, err
		} else {
			return float32(v), err
		}
	case dpb.FieldDescriptorProto_TYPE_INT64,
		dpb.FieldDescriptorProto_TYPE_SINT64,
		dpb.FieldDescriptorProto_TYPE_SFIXED64:
		if v, err := strconv.ParseInt(s, 10, 64); err != nil {
			return nil, err
		} else {
			return v, nil
		}
	case dpb.FieldDescriptorProto_TYPE_UINT64:
		if v, err := strconv.ParseUint(s, 10, 64); err != nil {
			return nil, err
		} else {
			return v, nil
		}
	case dpb.FieldDescriptorProto_TYPE_INT32,
		dpb.FieldDescriptorProto_TYPE_SINT32,
		dpb.FieldDescriptorProto_TYPE_SFIXED32,
		dpb.FieldDescriptorProto_TYPE_ENUM:
		if v, err := strconv.ParseInt(s, 10, 32); err != nil {
			return nil, err
		} else {
			return int32(v), nil
		}
	case dpb.FieldDescriptorProto_TYPE_UINT32:
		if v, err := strconv.ParseUint(s, 10, 32); err != nil {
			return nil, err
		} else {
			return uint32(v), nil
		}
	case dpb.FieldDescriptorProto_TYPE_BOOL:
		return len(s) != 0 && s != "0" && strings.ToLower(s) != "false", nil
	case dpb.FieldDescriptorProto_TYPE_STRING:
		return s, nil
	//case dpb.FieldDescriptorProto_TYPE_MESSAGE:
	default:
		return nil, fmt.Errorf("unsupported type: %v", t)
	}
}
