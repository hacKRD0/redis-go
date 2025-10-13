package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type handler interface {
	Handle(conn net.Conn, request string)
	handleKeys(conn net.Conn, args []string)
	handleSet(conn net.Conn, args []string)
	handleGet(conn net.Conn, args []string)
	handlePing(conn net.Conn, args []string)
	handleEcho(conn net.Conn, args []string)
	handleQuit(conn net.Conn, args []string)
}

type handlerImpl struct {
}

func (h *handlerImpl) Handle(conn net.Conn, request string) {
	_, cmd, parts := parseRequest(request)
	switch cmd {
	case "PING":
		h.handlePing(conn, parts)
	case "ECHO":
		h.handleEcho(conn, parts)
	case "SET":
		h.handleSet(conn, parts)
	case "GET":
		h.handleGet(conn, parts)
	case "CONFIG":
		h.handleConfig(conn, parts)
	case "KEYS":
		h.handleKeys(conn, parts)
	default:
		sendError(conn, "unknown command '"+cmd+"'")
	}
}

func (h *handlerImpl) handleKeys(conn net.Conn, parts []string) {
	fmt.Println("config", config)
	RDBfile, err := os.ReadFile(path.Join(config.rdb_dir, config.rdb_filename))
	if err != nil {
		respArrayEmpty := "*0\r\n"
		conn.Write([]byte(respArrayEmpty))
	}

	keysCommand := strings.ToLower(parts[0])
	if keysCommand == "*" {

		FBidx := bytes.Index(RDBfile, []byte{0xFB})
		keyStart := int(FBidx + 5)
		keyLength := int(RDBfile[FBidx+4])
		keyName := (string(RDBfile[keyStart : keyStart+keyLength]))
		respArrayKeyName := fmt.Sprintf("*1\r\n$%d\r\n%s\r\n", len(keyName), keyName)
		conn.Write([]byte(respArrayKeyName))
	}
}

func (h *handlerImpl) handleSet(conn net.Conn, parts []string) {

	key := parts[0]
	value := parts[1]

	currentTime := time.Now()

	px := -1
	isNX := false
	isXX := false
	i := 2
	for i < len(parts) {
		arg := strings.ToUpper(parts[i])
		switch arg {
		case "EX":
			if i+1 >= len(parts) {
				sendError(conn, "syntax error")
				return
			}
			secs, err := strconv.Atoi(parts[i+1])
			if err != nil {
				sendError(conn, "value is not an integer or out of range")
				return
			}
			px = secs * 1000
			i += 2
		case "PX":
			if i+1 >= len(parts) {
				sendError(conn, "syntax error")
				return
			}
			msecs, err := strconv.Atoi(parts[i+1])
			if err != nil {
				sendError(conn, "value is not an integer or out of range")
				return
			}
			px = msecs
			i += 2
		case "NX":
			isNX = true
			i++
		case "XX":
			isXX = true
			i++
		case "EXAT":
			if i+1 >= len(parts) {
				sendError(conn, "syntax error")
				return
			}
			unixSecs, err := strconv.ParseInt(parts[i+1], 10, 64)
			if err != nil {
				sendError(conn, "value is not an integer or out of range")
				return
			}
			px = int(unixSecs * 1000)
			i += 2
		case "PXAT":
			if i+1 >= len(parts) {
				sendError(conn, "syntax error")
				return
			}
			unixMsecs, err := strconv.ParseInt(parts[i+1], 10, 64)
			if err != nil {
				sendError(conn, "value is not an integer or out of range")
				return
			}
			px = int(unixMsecs)
			i += 2
		default:
			sendError(conn, "syntax error")
			return
		}
	}

	if isNX {
		_, ok := DB.LoadOrStore(key, kv{value, px, currentTime})
		if ok {
			sendSimpleString(conn, "OK")
			return
		}
	}

	if isXX {
		_, ok := DB.Load(key)
		if !ok {
			sendSimpleString(conn, "OK")
			return
		}
		DB.Store(key, kv{value, px, currentTime})
	} else {
		DB.Store(key, kv{value, px, currentTime})
	}

	sendSimpleString(conn, "OK")
}

func (h *handlerImpl) handleGet(conn net.Conn, parts []string) {
	if len(parts) != 1 {
		sendError(conn, "wrong number of arguments for 'GET'")
		return
	}

	key := parts[0]
	currentTime := time.Now()
	val, ok := DB.Load(key)
	if !ok {
		sendBulkString(conn, "")
		return
	}
	kvValue := val.(kv)
	if kvValue.px != -1 && currentTime.After(kvValue.t.Add(time.Duration(kvValue.px)*time.Millisecond)) {
		DB.Delete(key)
		sendBulkString(conn, "")
		return
	}
	sendBulkString(conn, kvValue.value)
}

func (h *handlerImpl) handlePing(conn net.Conn, parts []string) {
	sendSimpleString(conn, "PONG")
}

func (h *handlerImpl) handleEcho(conn net.Conn, parts []string) {
	if len(parts) == 0 {
		sendError(conn, "wrong number of arguments for 'ECHO'")
		return
	}
	message := strings.Join(parts, " ")
	sendBulkString(conn, message)
}

func (h *handlerImpl) handleQuit(conn net.Conn, parts []string) {
}

func (h *handlerImpl) handleConfig(conn net.Conn, parts []string) {
	if len(parts) < 2 {
		sendError(conn, "wrong number of arguments for 'CONFIG'")
		return
	}
	subcommand := strings.ToUpper(parts[0])
	switch subcommand {
	case "GET":
		if len(parts) != 2 {
			sendError(conn, "wrong number of arguments for 'CONFIG GET'")
			return
		}
		pattern := parts[1]
		var response []string
		if pattern == "*" || pattern == "dir" {
			response = append(response, "dir", config.rdb_dir)
		}
		if pattern == "*" || pattern == "dbfilename" {
			response = append(response, "dbfilename", config.rdb_filename)
		}
		fmt.Println("config", config)
		fmt.Println("response", response)
		fmt.Println("encodeArray(response)", encodeArray(response))
		conn.Write(encodeArray(response))
	default:
		sendError(conn, "unknown subcommand for 'CONFIG'")
	}
}
