package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func RedisSimpleString(s string) string {
	return fmt.Sprintf("+%s\r\n", s)
}

func RedisError(kind string, msg string) string {
	return fmt.Sprintf("-%s %s\r\n", kind, msg)
}

func RedisInteger(i int) string {
	return fmt.Sprintf(":%d\r\n", i)
}

type DataAccess struct {
	operation string
	key string
	value string
	responseCh chan string
}

type EchoCommand struct {
	value string
}

type SetCommand struct {
	key string
	value string
}

type GetCommand struct {
	key string
}

type PingCommand struct {}

type UnknownCommand struct {}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	fmt.Println("Listening on port 6379...")

	clientData := make(chan DataAccess)

	go manageData(clientData)

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
		}

		fmt.Println("Received request...")

		go handleConn(c, clientData)
	}
}

func handleConn(c net.Conn, ch chan<- DataAccess) {
	defer closeConn(c)
	
	for {
		requestBuffer := bufio.NewScanner	(c)

		tokenizer := Tokenizer{
			rawRequest: requestBuffer,
		}
		tokens, err := tokenizer.Tokenize()
		if errors.Is(err, io.EOF) {
			break
		}

		var command interface{}
		commandType := strings.ToUpper(tokens[0].subTokens[0].value)
		if commandType == "ECHO" {
			command = EchoCommand{value: tokens[0].subTokens[1].value}
		} else if commandType == "SET" {
			command = SetCommand{key: tokens[0].subTokens[1].value, value: tokens[0].subTokens[2].value}
		} else if commandType == "GET" {
			command = GetCommand{key: tokens[0].subTokens[1].value}
		} else if commandType == "PING" {
			command = PingCommand{}
		} else {
			command = UnknownCommand{}
		}

		switch resolvedCommand := command.(type) {
		case EchoCommand:
			c.Write([]byte(RedisSimpleString(resolvedCommand.value)))
		case SetCommand:
			responseCh := make(chan string)
			ch <- DataAccess{
				operation: "write", 
				key: resolvedCommand.key, 
				value: resolvedCommand.value, 
				responseCh: responseCh,
			}
			response := <-responseCh
			c.Write([]byte(RedisSimpleString(response)))
		case GetCommand:
			responseCh := make(chan string)
			ch <- DataAccess{
				operation: "read", 
				key: resolvedCommand.key, 
				responseCh: responseCh,
			}
			response := <-responseCh
			c.Write([]byte(RedisSimpleString(response)))
		case PingCommand:
			c.Write([]byte(RedisSimpleString("PONG")))
		case UnknownCommand:
			c.Write([]byte(RedisSimpleString("")))
		}
	}
}

func closeConn(c net.Conn) {
	fmt.Println("closing connection")
	c.Close()
}

func manageData(ch <-chan DataAccess) {
	dataStore := map[string]string{}

	for v := range(ch) {
		switch v.operation {
		case "write":
			dataStore[v.key] = v.value
			v.responseCh <- "OK"
		case "read":
			v.responseCh <- dataStore[v.key]
		}
	}
}
