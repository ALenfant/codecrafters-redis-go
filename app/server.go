package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ALenfant/codecrafters-redis-go/app/parser"
	"github.com/ALenfant/codecrafters-redis-go/app/store"
)

const NullValue string = "$-1\r\n"

type RedisServer struct {
	store *store.DataStore
}

func NewRedisServer() *RedisServer {
	return &RedisServer{
		store: store.NewDataStore(),
	}
}

func (s *RedisServer) Start() {
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
		go s.handleRequest(conn)
	}
}

func (s *RedisServer) handleRequest(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		parsedData, err := parser.ParseData(reader)
		if err != nil {
			conn.Close()
			log.Printf("Error parsing data: %v\n", err)
			return
		}

		switch parsedData := parsedData.(type) {
		case *parser.RedisArray:
			command := parsedData.Items[0]
			arguments := parsedData.Items[1:]
			switch command := command.(type) {
			case *parser.RedisBulkString:
				if err := s.runCommand(command.Content, arguments, conn); err != nil {
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

func (s *RedisServer) runCommand(command string, arguments []parser.RedisData, conn net.Conn) error {
	command = strings.ToLower(command)
	if command == "ping" {
		conn.Write([]byte("+PONG\r\n"))
	} else if command == "echo" {
		messageString, ok := arguments[0].(*parser.RedisBulkString)
		if !ok {
			return fmt.Errorf("expected bulkstring, got : %v", arguments[0])
		}
		conn.Write([]byte("+"))
		conn.Write([]byte(messageString.Content))
		conn.Write([]byte("\r\n"))
	} else if command == "set" {
		key, ok := arguments[0].(*parser.RedisBulkString)
		if !ok {
			return fmt.Errorf("expected key bulkstring, got : %v", arguments[0])
		}
		val, ok := arguments[1].(*parser.RedisBulkString)
		if !ok {
			return fmt.Errorf("expected val bulkstring, got : %v", arguments[0])
		}
		fmt.Printf("DDDEBUG %v\n", toJson(arguments))
		if len(arguments) > 2 {
			parameter := arguments[2].(*parser.RedisBulkString)
			parameterVal := arguments[3].(*parser.RedisBulkString)

			if strings.ToLower(parameter.Content) == "px" {
				expirationMs, err := strconv.ParseInt(parameterVal.Content, 10, 64)
				if err != nil {
					return fmt.Errorf("couldnt parse expiration ms %v : %v", *parameterVal, err)
				}
				expirationTime := time.Now().Add(time.Duration(expirationMs) * time.Millisecond)
				s.store.SetWithExpiration(key.Content, val.Content, &expirationTime)
			} else {
				return fmt.Errorf("unrecognized param %v %v", *parameter, *parameterVal)
			}
		} else {
			s.store.Set(key.Content, val.Content)
		}
		conn.Write([]byte("+OK\r\n"))
	} else if command == "get" {
		key, ok := arguments[0].(*parser.RedisBulkString)
		if !ok {
			return fmt.Errorf("expected key bulkstring, got : %v", arguments[0])
		}
		val := s.store.Get(key.Content)
		if val == nil {
			conn.Write([]byte(NullValue))
		} else {
			conn.Write([]byte("$"))
			conn.Write([]byte(fmt.Sprintf("%d", len(*val))))
			conn.Write([]byte("\r\n"))
			conn.Write([]byte(*val))
			conn.Write([]byte("\r\n"))
		}

	} else {
		return fmt.Errorf("unknown command: %s", command)
	}
	return nil
}

func toJson(obj interface{}) string {
	str, err := json.Marshal(obj)
	if err != nil {
		log.Panicf(fmt.Sprintf("%v", err))
	}
	return string(str)
}

func main() {
	server := NewRedisServer()
	server.Start()
}
