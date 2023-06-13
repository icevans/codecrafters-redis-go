package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func RedisSimpleString(s string) string {
	return fmt.Sprintf("+%s", s)
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
	for {
		r := make([]byte, 256)

		n, err := c.Read(r)
		
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println("error reading client connection: ", err.Error())
				os.Exit(1)
			}
		}

		r = r[:n]

		for _, b := range(r) {
			if b == 13 {
				fmt.Print("\\r")
			} else if b == 10 {
				fmt.Print("\\n")
			} else {
				fmt.Printf("%v", string([]byte{b}))
			}
		}
		fmt.Print("\n")

		tokenizer := Tokenizer{
			cursor: 0,
			str: string(r[:]),
		}

		tokens, _ := tokenizer.Tokenize()

		parser := RequestParser{
			cursor: -1,
			tokens: tokens,
		}

		command, _ := parser.Parse()

		if command.name == "ECHO" {
			c.Write([]byte(RedisSimpleString(command.inputs[0])))
		} else if command.name == "PING" {
			c.Write([]byte(RedisSimpleString("PONG")))
		} else {
			c.Write([]byte(RedisSimpleString("")))
		}
	}

	c.Close()
}
