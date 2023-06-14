package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type Token struct {
	kind  string
	value string
	subTokens []Token
}

type Tokenizer struct {
	rawRequest *bufio.Scanner
}

func (t *Tokenizer) getNextToken() (Token, error) {
	// TODO: reimplement this method with bufio.Scanner.Scan
	foundByte := t.rawRequest.Scan()
	if !foundByte {
		return Token{}, io.EOF
	}

	descriptor := t.rawRequest.Text()
	RESPType := descriptor[0]
	bulkLength, err := strconv.Atoi(descriptor[1:])
	if err != nil {
		return Token{}, fmt.Errorf("%v is not a valid bulk length", descriptor[1:])
	}

	if RESPType == '*' {
		return t.parseBulkArray(bulkLength)
	} else if RESPType == '$' {
		return t.parseBulkString(bulkLength)
	} else {
		return Token{}, errors.New("invalid RESP data type")
	}
}

func (t *Tokenizer) parseBulkArray(bulkLength int) (Token, error) {
	var arrayContents []Token
	for i := 0; i < bulkLength; i++ {
		arrayElement, _ := t.getNextToken()
		arrayContents = append(arrayContents, arrayElement)
	}
	
	return Token{
		kind: "BulkArray",
		subTokens: arrayContents,
	}, nil
}

func (t *Tokenizer) parseBulkString(bulkLength int) (Token, error) {
	t.rawRequest.Scan()

	return Token{
		kind: "BulkString",
		value: t.rawRequest.Text(),
	}, nil
}

func (t *Tokenizer) Tokenize() ([]Token, error) {
	var tokens []Token

	token, err := t.getNextToken()
	if err != nil {
		return nil, err
	}
	tokens = append(tokens, token)

	return tokens, nil
}
