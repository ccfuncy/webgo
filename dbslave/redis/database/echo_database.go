package database

import (
	databaseface "redis/interface/database"
	"redis/interface/resp"
	"redis/resp/reply"
)

// 回发内核层
type EchoDatabase struct {
}

func (e EchoDatabase) Exec(client resp.Connection, args databaseface.Cmdline) resp.Reply {
	bulkReply := reply.NewMultiBulkReply(args)
	return bulkReply
}

func (e EchoDatabase) Close() {

}

func (e EchoDatabase) AfterClose(client resp.Connection) {

}

func NewEchoDatabase() *EchoDatabase {
	return &EchoDatabase{}
}
