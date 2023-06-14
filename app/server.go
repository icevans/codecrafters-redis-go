package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
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

type EchoCommand struct {
	value string
}

type SetCommand struct {
	key string
	value string
	expiry int64
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

		// TODO: Move the raw call to the tokenizer and then the construction of 
		//       a command struct to its own CommandParser interface
		tokenizer := Tokenizer{
			rawRequest: requestBuffer,
		}
		tokens, err := tokenizer.Tokenize()
		if errors.Is(err, io.EOF) {
			break
		}

		var command interface{}
		commandType := strings.ToUpper(tokens[0].subTokens[0].value)
		switch commandType {
		case "ECHO":
			command = EchoCommand{
				value: tokens[0].subTokens[1].value,
			}
		case "SET":
			var expiry int64
			// TODO: This is gross, but let's get it working before cleaning up
			if len(tokens[0].subTokens) > 3 && strings.ToUpper(tokens[0].subTokens[3].value) == "PX" {
				expiryInt, err := strconv.Atoi(tokens[0].subTokens[4].value)
				if err != nil {
					fmt.Println("invalid expiry")
					break
				}
				expiry = int64(expiryInt)
			}

			command = SetCommand{
				key: tokens[0].subTokens[1].value, 
				value: tokens[0].subTokens[2].value,
				expiry: expiry, // will treat the 0 value of int64 as not setting an expiry
			}
		case "GET":
			command = GetCommand{
				key: tokens[0].subTokens[1].value,
			}
		case "PING":
			command = PingCommand{}
		default:
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
				expiry: resolvedCommand.expiry,
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
			c.Write([]byte(RedisBulkString(response)))
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

func manageData(ch chan DataAccess) {
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

			if v.expiry > 0 {
				go expireKey(v.key, v.expiry, ch)
			}
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
