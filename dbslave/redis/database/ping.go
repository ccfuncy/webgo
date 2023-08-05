package database

import (
	"redis/interface/database"
	"redis/interface/resp"
	"redis/resp/reply"
)

func init() {
	RegisterCommand("ping", Ping, 1)
}

func Ping(db *DB, cmdline database.Cmdline) resp.Reply {
	return reply.NewPongReply()
}
