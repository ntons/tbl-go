package tbl

import (
	"fmt"
	"reflect"

	pb "github.com/golang/protobuf/proto"
)

// field index for sheet
type index struct {
	t reflect.Type
	m map[interface{}]pb.Message
}

func newIndex(t reflect.Type) *index {
	return &index{t: t, m: make(map[interface{}]pb.Message)}
}
func (idx *index) check(key interface{}) (interface{}, error) {
	if t := reflect.TypeOf(key); idx.t != t {
		if !t.ConvertibleTo(idx.t) {
			return nil, fmt.Errorf("invalid key type")
		}
		key = reflect.ValueOf(key).Convert(idx.t).Interface()
	}
	return key, nil
}
func (idx *index) add(key interface{}, msg pb.Message) (err error) {
	if key, err = idx.check(key); err != nil {
		return
	}
	idx.m[key] = msg
	return
}
func (idx *index) get(key interface{}) (msg pb.Message, err error) {
	if key, err = idx.check(key); err != nil {
		return
	}
	var ok bool
	if msg, ok = idx.m[key]; !ok {
		return nil, fmt.Errorf("no such key")
	}
	return
}
