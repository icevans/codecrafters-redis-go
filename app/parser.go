package main

import (
	"errors"
	"strconv"
)

var validCommands = map[string]struct{}{
	"ECHO": {},
}

type RequestParser struct {
	cursor int
	tokens []Token
}

type Command struct {
	name string
	inputs []string
}

func (rp *RequestParser) Parse() (Command, error) {
	var request []string

	requestArrayDescriptor, err := rp.getNextElement()
	if err != nil {
		return Command{}, errors.New("no tokens")
	}

	requestLength, _ := strconv.Atoi(requestArrayDescriptor.value)

	for {
		requestPart, err := rp.getNextElement()
		if err != nil {
			break
		}
		request = append(request, requestPart.value)
	}

	if len(request) != requestLength {
		return Command{}, errors.New("incorrect number of elements in input")
	}

	if _, ok := validCommands[request[0]]; !ok {
		return Command{}, errors.New("unknown command")
	}

	return Command{
		name: request[0],
		inputs: request[1:],
	}, nil
}

func (rp *RequestParser) getNextElement() (Token, error) {
	rp.cursor++

	if (rp.cursor >= len(rp.tokens)) {
		return Token{}, errors.New("done parsing")
	}

	token := rp.tokens[rp.cursor]

	if token.kind == "BulkStringDescriptor" {
		expectedLength, _ := strconv.Atoi(token.value)

		token, _ = rp.getNextElement()

		if token.kind != "String" {
			return Token{}, errors.New("BulkStringDescriptor not followed by string")
		}

		if len(token.value) != expectedLength {
			return Token{}, errors.New("BulkString did not match specified length")
		}
	}

	
	return token, nil
}