package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
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
	defer c.Close()
	
	for {
		requestBuffer := bufio.NewReader(c)

		tokenizer := Tokenizer{
			rawRequest: requestBuffer,
		}

		tokens, _ := tokenizer.Tokenize()
		command := tokens[0].subTokens[0].value

		if command == "ECHO" {
			c.Write([]byte(RedisSimpleString(tokens[0].subTokens[1].value)))
		} else if command == "PING" {
			c.Write([]byte(RedisSimpleString("PONG")))
		} else {
			c.Write([]byte(RedisSimpleString("")))
		}
	}
}
