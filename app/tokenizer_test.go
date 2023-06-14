package main

import (
	"bufio"
	"bytes"
	"reflect"
	"testing"
)

func TestTokenizer_tokenize_echo(t *testing.T) {
	request := "*2\r\n$4\r\nECHO\r\n$11\r\nhello world\r\n"
	tr := Tokenizer{
		rawRequest: bufio.NewScanner(bytes.NewBufferString(request)),
	}

	tokens, err := tr.Tokenize()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	expected := []Token{
		{
			kind: "BulkArray",
			subTokens: []Token{
				{kind: "BulkString", value: "ECHO"},
				{kind: "BulkString", value: "hello world"},
			},
		},
	}

	if !reflect.DeepEqual(tokens, expected) {
		t.Errorf("Got: %v, expected: %v", tokens, expected)
		t.Fail()
	}
}

func TestTokenizer_tokenize_echo_easy(t *testing.T) {
	request := "*1\r\n$4\r\nPING\r\n$"
	tr := Tokenizer{
		rawRequest: bufio.NewScanner(bytes.NewBufferString(request)),
	}

	tokens, err := tr.Tokenize()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	expected := []Token{
		{
			kind: "BulkArray",
			subTokens: []Token{
				{kind: "BulkString", value: "PING"},
			},
		},
	}

	if !reflect.DeepEqual(tokens, expected) {
		t.Errorf("Got: %v, expected: %v", tokens, expected)
		t.Fail()
	}
}
