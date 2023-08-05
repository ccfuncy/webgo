package database

import "redis/interface/resp"

type Cmdline [][]byte

type Database interface {
	Exec(client resp.Connection, args Cmdline) resp.Reply
	Close()
	AfterClose(client resp.Connection)
}

type DataEntity struct {
	Data interface{}
}
