package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

func RedisSimpleString(s string) string {
	return fmt.Sprintf("+%s\r\n", s)
}

func RedisBulkString(s string) string {
	if s == "" {
		return "$-1\r\n"
	}

	return fmt.Sprintf("$%v\r\n%v\r\n", len(s), s)
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
	expiry int64
	responseCh chan string
}

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
		command, err := parseCommand(c)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			fmt.Println(error.Error(err))
			c.Write([]byte(RedisSimpleString(error.Error(err))))
			continue
		}

		switch command.name {
		case "EchoCommand":
			c.Write([]byte(RedisSimpleString(command.input.value)))
		case "SetCommand":
			responseCh := make(chan string)
			ch <- DataAccess{
				operation: "write", 
				key: command.input.key, 
				value: command.input.value, 
				expiry: command.input.expiry,
				responseCh: responseCh,
			}
			if command.input.expiry > 0 {
				go expireKey(command.input.key, command.input.expiry, ch)
			}
			response := <-responseCh
			c.Write([]byte(RedisSimpleString(response)))
		case "GetCommand":
			responseCh := make(chan string)
			ch <- DataAccess{
				operation: "read",
				key: command.input.key, 
				responseCh: responseCh,
			}
			response := <-responseCh
			c.Write([]byte(RedisBulkString(response)))
		case "PingCommand":
			c.Write([]byte(RedisSimpleString("PONG")))
		case "UnknownCommand":
			c.Write([]byte(RedisSimpleString("")))
		}
	}
}

func closeConn(c net.Conn) {
	fmt.Println("closing connection")
	c.Close()
}

func manageData(ch <-chan DataAccess) {
	// TODO: Is this the most efficient data structure for a key value
	//       store? Read and write are both O(1), but there's no simple
	//       way to handle automatic eviction. We could consider an LRU
	//       cache instead
	dataStore := map[string]string{}

	for v := range(ch) {
		switch v.operation {
		case "write":
			dataStore[v.key] = v.value
			v.responseCh <- "OK"
		case "read":
			v.responseCh <- dataStore[v.key]
		case "delete":
			delete(dataStore, v.key)
		}
	}
}

func expireKey(key string, expiry int64, ch chan<- DataAccess) {
	expiryStart := time.Now()
	for {
		if time.Since(expiryStart).Milliseconds() >= expiry {
			ch <- DataAccess{
				operation: "delete",
				key: key,
			}
			break
		}
	}
}
