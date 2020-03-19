package tbl

import (
	"reflect"
	"testing"

	example "github.com/ntons/tbl-go/example/proto"
)

func TestLoad(t *testing.T) {
	var v example.ExampleSheet1
	s, err := LoadFile(reflect.TypeOf(v), "example/tbl/example_sheet1.tbl")
	if err != nil {
		t.Fatalf("load file fail: %v", err)
	}
	if err := s.CreateIndex("Id"); err != nil {
		t.Fatalf("create index fail: %v", err)
	}
	if _, err = s.FindByIndex("Id", 100002); err != nil {
		t.Fatalf("get by index fail: %v", err)
	}
	if _, err = s.FindByIndex("Id", "200001"); err == nil {
		t.Fatalf("unexpected err returned")
	}
}
