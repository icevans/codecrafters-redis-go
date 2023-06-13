package main

import (
	"bufio"
	"errors"
	"fmt"
	"strconv"
)

type Token struct {
	kind  string
	value string
	subTokens []Token
}

type Tokenizer struct {
	rawRequest *bufio.Reader
}

func (t *Tokenizer) getNextToken() (Token, error) {
	// TODO: reimplement this method with bufio.Scanner.Scan
	RESPType, err := t.rawRequest.ReadByte()
	if err != nil {
		return Token{}, err
	}

	// TODO: rename bulkLengthBytes to bulkLengthBytes when moving to recursive lexer/parser
	var bulkLengthBytes []byte
	nextByte, err := t.rawRequest.ReadByte()
	if err != nil {
		return Token{}, err
	}

	for nextByte != '\r' {
		bulkLengthBytes = append(bulkLengthBytes, nextByte)
		nextByte, err = t.rawRequest.ReadByte()
		if err != nil {
			// unexpected, ended up with the rest of the string
			return Token{}, err
		}
	}

	t.rawRequest.ReadByte() // Skip past the next \n

	bulkLength, err := strconv.Atoi(string(bulkLengthBytes))
	if err != nil {
		return Token{}, fmt.Errorf("%v is not a valid bulk length", string(bulkLengthBytes))
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
		var stringBytes []byte
		for i := 0; i < bulkLength; i++ {
			stringByte, _ := t.rawRequest.ReadByte()
			stringBytes = append(stringBytes, stringByte)
		}

		t.rawRequest.Discard(2)

		return Token{
			kind: "BulkString",
			value: string(stringBytes),
		}, nil
	} else {
		return Token{}, errors.New("invalid RESP data type")
	}
}

func (t *Tokenizer) Tokenize() ([]Token, error) {
	var tokens []Token

	token, _ := t.getNextToken()
	tokens = append(tokens, token)

	return tokens, nil
}
