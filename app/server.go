package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
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
		message, err := reader.ReadString('\r')
		if err != nil {
			conn.Close()
			log.Printf("Error reading message: %v\n", err)
			return
		}
		if len(message) < 1 {
			conn.Close()
			log.Printf("Empty message\n")
			return
		}

		fmt.Printf("Message incoming: %s", string(message))

		valueType := message[0]
		if valueType == '*' {
			arrayLength := message[1:]
			fmt.Printf("ARRAY LENGTH %v", arrayLength)
		} else {
			conn.Close()
			log.Printf("Unknown type: %v\n", valueType)
			return
		}
		conn.Write([]byte("+PONG\r\n"))
	}
}
