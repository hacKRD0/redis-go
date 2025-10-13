package main

import (
	"time"
)

type kv struct {
	value string
	px    int
	t     time.Time
}

type Config struct {
	// Port         string
	// Role         string
	// MasterHost   string
	// MasterPort   string
	// replicaConns map[string]net.Conn
	// ReplicaMu    sync.Mutex
	// ReplOffset   int64
	// ReplicaAcks  map[string]int64
	rdb_dir      string
	rdb_filename string
}
