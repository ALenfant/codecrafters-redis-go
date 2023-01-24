package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/ALenfant/codecrafters-redis-go/app/parser"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		parsedData, err := parser.ParseData(reader)
		if err != nil {
			conn.Close()
			log.Printf("Error parsing data: %v\n", err)
			return
		}
		fmt.Printf("%v", parsedData)

		switch parsedData := parsedData.(type) {
		case *parser.RedisArray:
			command := parsedData.Items[0]
			arguments := parsedData.Items[1:]
			switch command := command.(type) {
			case *parser.RedisBulkString:
				if err := runCommand(command.Content, arguments, conn); err != nil {
					log.Printf("Error while running command: %#v: %v\n", *command, err)
					return
				}
			default:
				conn.Close()
				log.Printf("expected bulkstring command, but received: %#v\n", command)
				return
			}

		default:
			conn.Close()
			log.Printf("expected array of commands, but received: %#v\n", parsedData)
			return
		}
	}
}

func runCommand(command string, arguments []parser.RedisData, conn net.Conn) error {
	command = strings.ToLower(command)
	if command == "ping" {
		conn.Write([]byte("+PONG\r\n"))
	} else if command == "echo" {
		log.Printf("DDDDDDDDDD: %#v\n", arguments)
		messageString, ok := arguments[0].(*parser.RedisBulkString)
		if !ok {
			return fmt.Errorf("expected bulkstring, got : %v", arguments[0])
		}
		conn.Write([]byte("+"))
		conn.Write([]byte(messageString.Content))
		conn.Write([]byte("\r\n"))
	} else {
		return fmt.Errorf("unknown command: %s", command)
	}
	return nil
}
