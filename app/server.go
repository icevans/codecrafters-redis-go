package main

import (
	"fmt"
	"io"
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

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
		}

		go handleConn(c)
	}
}

func handleConn(c net.Conn) {
	for {
		r := make([]byte, 256)

		if _, err := c.Read(r); err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println("error reading client connection: ", err.Error())
				os.Exit(1)
			}
		}

		fmt.Println(string(r[:]))

		c.Write([]byte(RedisSimpleString("PONG")))
	}

	c.Close()
}

/*
Parse command. Assumptions:
1. An input is always a RESP array containing RESP bulk strings
2. The command name will be the first element in the array
3. An input only has one command

- getNextToken... first time it should return array descriptor or error!
- loop number of times from array descriptor
	- getNextToken... it better be bulk string descriptor or error!
	- getNextByteString with length from bulk string descriptor

-
*/
