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
