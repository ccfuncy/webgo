package database

import (
	"redis/interface/database"
	"redis/interface/resp"
	"redis/interface/utils"
	"redis/resp/reply"
)

func init() {
	RegisterCommand("get", execGet, 2)
	RegisterCommand("set", execSet, 3)
	RegisterCommand("setnx", execSetNx, 3)
	RegisterCommand("getset", execGetSet, 3)
	RegisterCommand("strlen", execStrlen, 2)
}

// Get k1
func execGet(db *DB, args database.Cmdline) resp.Reply {
	key := utils.BytesToString(args[0])
	entity, exist := db.GetEntity(key)
	if !exist {
		return reply.NewEmptyBulkReply()
	}
	//todo: 本次实现只支持string数据类型
	bytes := entity.Data.([]byte)
	return reply.NewBulkReply(bytes)
}

// Set k v
func execSet(db *DB, args database.Cmdline) resp.Reply {
	key := utils.BytesToString(args[0])
	value := args[1]
	db.PutEntity(key, &database.DataEntity{Data: value})
	return reply.NewOKReply()
}

// SetNx k v
func execSetNx(db *DB, args database.Cmdline) resp.Reply {
	key := utils.BytesToString(args[0])
	value := args[1]
	result := db.PutIfAbsent(key, &database.DataEntity{Data: value})
	return reply.NewIntReply(int64(result))
}

// GetSet k1 v1
func execGetSet(db *DB, args database.Cmdline) resp.Reply {
	key := utils.BytesToString(args[0])
	newValue := args[1]
	entity, exist := db.GetEntity(key)
	db.PutEntity(key, &database.DataEntity{Data: newValue})
	if !exist {
		return reply.NewNULLBulkReply()
	}
	return reply.NewBulkReply(entity.Data.([]byte))
}

// Strlen k
func execStrlen(db *DB, args database.Cmdline) resp.Reply {
	key := utils.BytesToString(args[0])
	entity, exist := db.GetEntity(key)
	if !exist {
		return reply.NewNULLBulkReply()
	}
	return reply.NewIntReply(int64(len(entity.Data.([]byte))))
}
