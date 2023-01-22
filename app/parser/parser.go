package parser

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type RedisType int

const RedisTypeArray = 1
const RedisTypeBulkString = 2

type RedisTypeInfo struct {
	Type   RedisType
	Length int
}

type RedisArray struct {
	Items []interface{}
}

type RedisBulkString string

func ParseData(reader *bufio.Reader) (interface{}, error) {
	message, err := reader.ReadString('\r')
	if err != nil {
		return nil, fmt.Errorf("error reading message: %v", err)
	}
	nextByte, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("error reading next byte: %v", err)
	}
	if nextByte != '\n' {
		return nil, fmt.Errorf("wrong next byte message: %v", nextByte)
	}
	message = message[:len(message)-1] // Remove the trailing \r
	if len(message) < 1 {
		return nil, fmt.Errorf("empty message")
	}

	fmt.Printf("DEBUG Message incoming: %s\n", string(message))

	redisType, err := parseType(message)
	if err != nil {
		return nil, fmt.Errorf("message parsing error: %v", err)
	}

	if redisType.Type == RedisTypeArray {
		items, err := parseArrayItems(reader, redisType.Length)
		if err != nil {
			return nil, fmt.Errorf("array parsing error: %v", err)
		}
		return &RedisArray{items}, nil
	} else if redisType.Type == RedisTypeBulkString {
		content, err := parseBulkString(reader, redisType.Length)
		if err != nil {
			return nil, fmt.Errorf("bulk string parsing error: %v", err)
		}
		contentBulkString := RedisBulkString(content)
		return &contentBulkString, nil
	} else {
		return nil, fmt.Errorf("unsupported type: %#v", redisType)
	}
}

func parseType(message string) (*RedisTypeInfo, error) {
	valueType := message[0]
	if valueType == '*' { // ARRAY
		arrayLength, err := strconv.Atoi(message[1:])
		if err != nil {
			return nil, err
		}
		fmt.Printf("ARRAY LENGTH %v\n", arrayLength)
		return &RedisTypeInfo{Type: RedisTypeArray, Length: arrayLength}, nil
	} else if valueType == '$' { //Bulk String
		stringLength, err := strconv.Atoi(message[1:])
		if err != nil {
			return nil, err
		}
		fmt.Printf("STRING LENGTH %v\n", stringLength)
		return &RedisTypeInfo{Type: RedisTypeBulkString, Length: stringLength}, nil
	} else {
		return nil, fmt.Errorf("unknown type: %c (%v)", valueType, valueType)
	}
}

func parseArrayItems(reader *bufio.Reader, arrayLength int) ([]interface{}, error) {
	var arrayItems []interface{}
	for i := 0; i < arrayLength; i++ {
		item, err := ParseData(reader)
		if err != nil {
			return nil, fmt.Errorf("error parsing array iten at index %v: %v", i, err)
		}
		arrayItems = append(arrayItems, item)
	}
	return arrayItems, nil
}

func parseBulkString(reader *bufio.Reader, stringLength int) (string, error) {
	contentBytes := make([]byte, stringLength)
	_, err := io.ReadFull(reader, contentBytes)
	if err != nil {
		return "", fmt.Errorf("error parsing bulk string: %v", err)
	}
	reader.Discard(2) // Skip trailing \r\n
	return string(contentBytes), nil
}
