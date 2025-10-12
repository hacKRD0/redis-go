package main

import (
	"time"
)

type kv struct {
	value string
	px    int
	t     time.Time
}

type Cmd struct {
	cmd  string
	args []string
}

type RdbConfig struct {
	dir        string
	dbfilename string
}
