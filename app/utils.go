package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

var DB sync.Map
var config Config

func parseRequest(request string) (string, string, []string) {
	args := strings.Split(request, "\r\n")
	noOfElements := args[0][1:]
	command := strings.ToUpper(args[2])
	parts := []string{}
	for i := 4; i < len(args); i += 2 {
		parts = append(parts, args[i])
	}
	return noOfElements, command, parts
}

// Helper functions for response formatting
func sendError(conn net.Conn, message string) {
	response := fmt.Sprintf("-ERR %s\r\n", message)
	conn.Write([]byte(response))
}

func sendSimpleString(conn net.Conn, message string) {
	response := fmt.Sprintf("+%s\r\n", message)
	conn.Write([]byte(response))
}

func sendBulkString(conn net.Conn, value string) {
	if value == "" {
		conn.Write([]byte("$-1\r\n"))
		return
	}
	response := fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
	conn.Write([]byte(response))
}

func encodeArray(items []string) []byte {
	var parts []string
	for _, item := range items {
		parts = append(parts, fmt.Sprintf("$%d\r\n%s\r\n", len(item), item))
	}
	response := fmt.Sprintf("*%d\r\n%s", len(parts), strings.Join(parts, ""))
	return []byte(response)
}
