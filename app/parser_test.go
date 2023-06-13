package main

import (
	"reflect"
	"testing"
)

func TestParser_parse(t *testing.T) {
	rp := RequestParser{
		cursor: -1,
		tokens: []Token{
			{kind: "ArrayDescriptor", value: "3"},
			{kind: "BulkStringDescriptor", value: "4"},
			{kind: "String", value: "ECHO"},
			{kind: "BulkStringDescriptor", value: "5"},
			{kind: "String", value: "hello"},
			{kind: "BulkStringDescriptor", value: "5"},
			{kind: "String", value: "world"},
		},
	}

	request, _ := rp.Parse()
	expected := Command{
		name: "ECHO",
		inputs: []string{"hello", "world"},
	}

	if !reflect.DeepEqual(request, expected) {
		t.Fail()
	}
}