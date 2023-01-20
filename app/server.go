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
		log.Printf("Error dddfd")

		parsedData, err := parser.ParseData(reader)
		if err != nil {
			conn.Close()
			log.Printf("Error parsing data: %v\n", err)
			return
		}
		fmt.Printf("%v", parsedData)

		switch parsedData := parsedData.(type) {
		case *parser.RedisArray:
			for _, command := range parsedData.Items {
				switch command := command.(type) {
				case *parser.RedisBulkString:
					runCommand(*command, conn)
				default:
					conn.Close()
					log.Printf("expected bulkstring command, but received: %#v\n", command)
					return
				}
			}

		default:
			conn.Close()
			log.Printf("expected array of commands, but received: %#v\n", parsedData)
			return
		}
	}
}

func runCommand(command parser.RedisBulkString, conn net.Conn) {
	commandString := strings.ToLower(string(command))
	if commandString == "ping" {
		conn.Write([]byte("+PONG\r\n"))
	}
}
