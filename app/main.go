package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

var store = make(map[string]string)

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
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	for {
		buffer := make([]byte, 1024)
		_, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			conn.Close()
			return
		}

		_, command, args := parseRequest(string(buffer))
		switch command {
		case "PING":
			conn.Write([]byte("+PONG\r\n"))
		case "ECHO":
			str := strings.Join(args, " ")
			resp := fmt.Sprintf("$%d\r\n%s\r\n", len(str), str)
			conn.Write([]byte(resp))
		case "SET":
			key := args[0]
			value := args[1]
			store[key] = value
			conn.Write([]byte("+OK\r\n"))
		case "GET":
			key := args[0]
			value, ok := store[key]
			if !ok {
				conn.Write([]byte("-1\r\n"))
				return
			}
			conn.Write([]byte("+" + value + "\r\n"))
		default:
			conn.Write([]byte("-ERR unknown command '" + command + "'\r\n"))
		}
	}
}

func parseRequest(request string) (string, string, []string) {
	parts := strings.Split(request, "\r\n")
	noOfElements := parts[0][1:]
	command := parts[2]
	args := []string{}
	for i := 4; i < len(parts); i += 2 {
		args = append(args, parts[i])
	}
	return noOfElements, command, args
}
