package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Command struct {
	name string
	input CommandInput
}

type CommandInput struct {
	key string
	value string
	expiry int64
}

func parseCommand(r io.Reader) (Command, error) {
	commandByteScanner := bufio.NewScanner(r)
	tokenizer := Tokenizer{
		rawRequest: commandByteScanner,
	}

	// A command should only ever have one top level token, so only call getNextToken once
	rawCommand, err := tokenizer.getNextToken()
	if err != nil {
		return Command{}, err
	}

	// Validation
	if rawCommand.kind != "BulkArray" {
		return Command{}, fmt.Errorf("expected bulk array, got: %v", rawCommand.kind)
	}

	if len(rawCommand.subTokens) < 1 {
		return Command{}, errors.New("Command array must contain at least 1 element")
	}

	for _, token := range(rawCommand.subTokens) {
		if token.kind != "BulkString" {
			return Command{}, 
			fmt.Errorf("elements of BulkArray must be only BulkStrings, got: %v", token.kind)
		}
	}

	commandName := strings.ToUpper(rawCommand.subTokens[0].value)
	rawCommandInput := rawCommand.subTokens[1:]

	switch commandName {
	case "ECHO":
		commandInput, err := parseEchoInput(rawCommandInput)
		if err != nil {
			return Command{}, fmt.Errorf("invalid command input: %v", err.Error())
		}

		return Command{
			name: "EchoCommand",
			input: commandInput,
		}, nil
	case "SET":
		commandInput, err := parseSetInput(rawCommandInput)
		if err != nil {
			return Command{}, fmt.Errorf("invalid command input: %v", err.Error())
		}
		
		return Command{
			name: "SetCommand",
			input: commandInput,
		}, nil
	case "GET":
		commandInput, err := parseGetInput(rawCommandInput)
		if err != nil {
			return Command{}, fmt.Errorf("invalid command input: %v", err.Error())
		}

		return Command{
			name: "GetCommand",
			input: commandInput,
		}, nil
	case "PING":
		return Command{
			name:  "PingCommand",
		}, nil
	default:
		return Command{
			name: "UnknownCommand",
		}, nil
	}
}

func parseEchoInput(t []Token) (CommandInput, error) {
	if len(t) != 1 {
		return CommandInput{}, errors.New("wrong number of arguments for ECHO")
	}

	return CommandInput{
		value:  t[0].value,
	}, nil
}

func parseSetInput(t []Token) (CommandInput, error) {
	if len(t) < 2 {
		return CommandInput{}, errors.New("wrong number of arguments for SET")
	}

	// TODO: What we really need here is a command flag parser
	var expiry int64
	if len(t) > 2 {
		if len(t) != 4 || strings.ToUpper(t[2].value) != "PX" {
			return CommandInput{}, errors.New("ECHO supports one flag PX, which takes one value")
		}
		
		expiryInt, err := strconv.Atoi(t[3].value)
		if err != nil {
			return CommandInput{}, errors.New("invalid expiry")
		}
		expiry = int64(expiryInt)
	}

	return CommandInput{
		value:  t[1].value,
		expiry: expiry,
	}, nil
}

func parseGetInput(t []Token) (CommandInput, error) {
	if len(t) != 1 {
		return CommandInput{}, errors.New("wrong number of arguments for GET")
	}

	return CommandInput{
		value:  t[0].value,
	}, nil
}


