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
	bulkLengthString := descriptor[1:]
	bulkLength, err := strconv.Atoi(string(bulkLengthString))
	if err != nil {
		return Token{}, fmt.Errorf("%v is not a valid bulk length", bulkLengthString)
	}

	if RESPType == '*' {
		var arrayContents []Token
		for i := 0; i < bulkLength; i++ {
			arrayElement, _ := t.getNextToken()
			arrayContents = append(arrayContents, arrayElement)
		}
		return Token{
			kind: "BulkArray",
			subTokens: arrayContents,
		}, nil
	} else if RESPType == '$' {
		t.rawRequest.Scan()

		return Token{
			kind: "BulkString",
			value: t.rawRequest.Text(),
		}, nil
	} else {
		return Token{}, errors.New("invalid RESP data type")
	}
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
