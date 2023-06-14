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

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	fmt.Println("Listening on port 6379...")

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
		}

		fmt.Println("Received request...")

		go handleConn(c)
	}
}

func handleConn(c net.Conn) {
	defer closeConn(c)
	
	for {
		requestBuffer := bufio.NewScanner(c)

		tokenizer := Tokenizer{
			rawRequest: requestBuffer,
		}
		tokens, err := tokenizer.Tokenize()
		if errors.Is(err, io.EOF) {
			break
		}

		command := tokens[0].subTokens[0].value

		if strings.ToUpper(command) == "ECHO" {
			c.Write([]byte(RedisSimpleString(tokens[0].subTokens[1].value)))
		} else if strings.ToUpper(command) == "PING" {
			c.Write([]byte(RedisSimpleString("PONG")))
		} else {
			c.Write([]byte(RedisSimpleString("")))
		}
	}
}

func closeConn(c net.Conn) {
	fmt.Println("closing connection")
	c.Close()
}
