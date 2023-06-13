package main

import (
	"reflect"
	"testing"
)

func TestTokenizer_tokenize_echo(t *testing.T) {
	tr := Tokenizer{
		cursor: 0,
		str:    "*3\r\n$4\r\nECHO\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
	}

	tokens, _ := tr.Tokenize()

	expected := []Token{
		{kind: "ArrayDescriptor", value: "3"},
		{kind: "BulkStringDescriptor", value: "4"},
		{kind: "String", value: "ECHO"},
		{kind: "BulkStringDescriptor", value: "5"},
		{kind: "String", value: "hello"},
		{kind: "BulkStringDescriptor", value: "5"},
		{kind: "String", value: "world"},
	}

	if !reflect.DeepEqual(tokens, expected) {
		t.Errorf("Got: %v, expected: %v", tokens, expected)
		t.Fail()
	}
}
