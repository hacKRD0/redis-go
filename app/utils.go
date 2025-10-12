package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var DB sync.Map
var config RdbConfig

func parseTime(timeString string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", timeString)
	if err != nil {
		return time.Time{}
	}
	return t
}

func handleRequest(conn net.Conn, request string) {
	_, cmd := parseRequest(request)
	switch cmd.cmd {
	case "PING":
		conn.Write([]byte("+PONG\r\n"))
	case "ECHO":
		str := strings.Join(cmd.args, " ")
		resp := fmt.Sprintf("$%d\r\n%s\r\n", len(str), str)
		conn.Write([]byte(resp))
	case "SET":
		handleSet(cmd.args)
		conn.Write([]byte("+OK\r\n"))
	case "GET":
		value := handleGet(cmd.args)
		fmt.Println("value: ", value)
		if value == "" {
			conn.Write([]byte("$-1\r\n"))
			return
		}
		conn.Write([]byte("+" + value + "\r\n"))
	case "CONFIG":
		if cmd.args[0] == "GET" {
			bulkString := handleConfigGet(cmd.args)
			conn.Write([]byte(bulkString))
		}
	default:
		conn.Write([]byte("-ERR unknown command '" + cmd.cmd + "'\r\n"))
	}
}

func handleConfigGet(args []string) string {
	switch args[1] {
	case "dir":
		return fmt.Sprintf("*2\r\n$3\r\ndir\r\n$%d\r\n%s\r\n", len(config.dir), config.dir)
	case "dbfilename":
		return fmt.Sprintf("*2\r\n$9\r\ndbfilename\r\n$%d\r\n%s\r\n", len(config.dbfilename), config.dbfilename)
	default:
		return ""
	}
}

func handleSet(args []string) {
	key := args[0]
	value := args[1]

	currentTime := time.Now()
	fmt.Println(currentTime)
	fmt.Println(args)

	px := -1
	isNX := false
	isXX := false
	for i := range args {
		arg := strings.ToUpper(args[i])
		switch arg {
		case "EX":
			secs, err := strconv.Atoi(args[i+1])
			if err != nil {
				return
			}
			px = secs * 1000
		case "PX":
			msecs, err := strconv.Atoi(args[i+1])
			if err != nil {
				return
			}
			px = msecs
		case "NX":
			isNX = true
		case "XX":
			isXX = true
		case "EXAT":
			unixSecs, err := strconv.ParseInt(args[i+1], 10, 64)
			if err != nil {
				return
			}
			px = int(unixSecs * 1000)
		case "PXAT":
			unixMsecs, err := strconv.ParseInt(args[i+1], 10, 64)
			if err != nil {
				return
			}
			px = int(unixMsecs)
		}
	}

	if isNX {
		_, ok := DB.LoadOrStore(key, kv{value, px, currentTime})
		if ok {
			return
		}
	}

	if isXX {
		_, ok := DB.Load(key)
		if !ok {
			return
		}
		DB.Store(key, kv{value, px, currentTime})
	}

	DB.Store(key, kv{value, px, currentTime})
}

func handleGet(args []string) string {
	key := args[0]
	currentTime := time.Now()
	fmt.Println(currentTime)
	val, ok := DB.Load(key)
	fmt.Println(val)
	if !ok {
		fmt.Println("not found")
		return ""
	}
	kvValue := val.(kv)
	if kvValue.px != -1 && currentTime.After(kvValue.t.Add(time.Duration(kvValue.px)*time.Millisecond)) {
		DB.Delete(key)
		return ""
	}
	return kvValue.value
}

func parseRequest(request string) (string, Cmd) {
	parts := strings.Split(request, "\r\n")
	noOfElements := parts[0][1:]
	command := strings.ToUpper(parts[2])
	args := []string{}
	for i := 4; i < len(parts); i += 2 {
		args = append(args, parts[i])
	}
	return noOfElements, Cmd{command, args}
}

func loadRdb(cfg RdbConfig) {
	config = cfg
}
