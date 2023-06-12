package main

import (
	"errors"
)

type Token struct {
	kind  string
	value string
}

type Tokenizer struct {
	cursor int
	str    string
}

var kindLookup = map[string]string{
	"*": "ArrayDescriptor",
	"$": "BulkStringDescriptor",
}

func (t *Tokenizer) getNextToken() (Token, error) {
	if t.cursor >= len(t.str) {
		return Token{}, errors.New("no more to parse")
	}

	nextToken := []byte{}
	for string(t.str[t.cursor]) != "\r" {
		nextToken = append(nextToken, t.str[t.cursor])
		t.cursor++
	}

	t.cursor += 2 // Advance past next \r\n

	// Baby's first lexer...
	if kind, ok := kindLookup[string(nextToken[0])]; ok {
		return Token{
			kind: kind,
			value: string(nextToken[1:]),
		}, nil
	} else {
		return Token{
			kind: "String",
			value: string(nextToken[:]),
		}, nil
	}
}

func (t *Tokenizer) Tokenize() ([]Token, error) {
	var tokens []Token

	token, err := t.getNextToken()
	if err != nil {
		// TODO
	}
	tokens = append(tokens, token)

	for err == nil {
		token, err = t.getNextToken()
		if err == nil {
			tokens = append(tokens, token)
		}
	}

	if err.Error() != "no more to parse" {
		return nil, err
	}

	return tokens, nil
}
