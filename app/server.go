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
			// Note, according to the Go docs on the io.Reader interface,
			// even when an io.EOF error is thrown, the reader may have
			// read a non-zero number of bytes, and so those should be
			// considered before handling the error, so this is slightly
			// incorrect.
			if err == io.EOF {
				break
			} else {
				fmt.Println("error reading client connection: ", err.Error())
				os.Exit(1)
			}
		}

		c.Write([]byte(RedisSimpleString("PONG")))
	}
	c.Close()
}
