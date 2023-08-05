package database

import (
	"fmt"
	"redis/datastruct/dict"
	"redis/interface/database"
	"redis/interface/resp"
	"redis/interface/utils"
	"redis/lib/logger"
	"redis/resp/reply"
	"strings"
)

type ExecFunc func(db *DB, cmdline database.Cmdline) resp.Reply

type DB struct {
	index int
	data  dict.Dict
}

func NewDB() *DB {
	return &DB{
		data: dict.NewSyncDict(),
	}
}

func (d *DB) Exec(connection resp.Connection, cmdline database.Cmdline) resp.Reply {
	//Ping Set setnx
	name := strings.ToUpper(utils.BytesToString(cmdline[0]))
	logger.Default().Info(fmt.Sprintf("client %s exec %s", connection.RemoteAddr().String(), name))
	cmd, ok := cmdTable[name]
	if !ok {
		return reply.NewStandardErrReply("Err unknown command " + name)
	}
	if !validateArity(cmd.arity, cmdline) {
		return reply.NewArgNumErrReply(name)
	}
	// Set k v => k v
	executor := cmd.executor(d, cmdline[1:])
	return executor
}

// 两种情况
// 1. 定长 set k v arity=2
// 2. 变长 exist v1 v2 v3 ... arity=-3 负号表示能够超过
func validateArity(arity int, argsCmd database.Cmdline) bool {
	length := len(argsCmd)
	if arity >= 0 {
		return arity == length
	} else {
		return length >= -arity
	}
}

func (d *DB) GetEntity(key string) (*database.DataEntity, bool) {
	raw, ok := d.data.Get(key)
	if !ok {
		return nil, false
	}
	entity := raw.(*database.DataEntity)
	return entity, true
}

func (d *DB) PutEntity(key string, val *database.DataEntity) int {
	return d.data.Put(key, val)
}

func (d *DB) PutIfExist(key string, val *database.DataEntity) int {
	return d.data.PutIfExist(key, val)
}

func (d *DB) PutIfAbsent(key string, val *database.DataEntity) int {
	return d.data.PutIfAbsent(key, val)
}

func (d *DB) Remove(key string) int {
	return d.data.Remove(key)
}

func (d *DB) Removes(keys ...string) int {
	delete := 0
	for _, key := range keys {
		delete += d.data.Remove(key)
	}
	return delete
}

func (d *DB) Flush() {
	d.data.Clear()
}
