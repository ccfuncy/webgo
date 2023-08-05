package database

import (
	"redis/interface/database"
	"redis/interface/resp"
	"redis/interface/utils"
	"redis/interface/wildcard"
	"redis/resp/reply"
)

func init() {
	RegisterCommand("del", execDel, -2)
	RegisterCommand("exists", execExist, -2)
	RegisterCommand("flushdb", execFlushDB, -1)
	RegisterCommand("type", execType, 2)
	RegisterCommand("rename", execRename, 3)
	RegisterCommand("renamenx", execRenameNx, 3)
	RegisterCommand("keys", execKeys, 2)
}

// Del
func execDel(db *DB, args database.Cmdline) resp.Reply {
	keys := make([]string, len(args))
	for i, arg := range args {
		keys[i] = utils.BytesToString(arg)
	}
	removes := db.Removes(keys...)
	return reply.NewIntReply(int64(removes))
}

// Exist k1 k2 k3
func execExist(db *DB, args database.Cmdline) resp.Reply {
	res := 0
	for _, arg := range args {
		_, exist := db.GetEntity(utils.BytesToString(arg))
		if exist {
			res++
		}
	}
	return reply.NewIntReply(int64(res))
}

// FlushDB
func execFlushDB(db *DB, args database.Cmdline) resp.Reply {
	db.Flush()
	return reply.NewOKReply()
}

// Type k1
func execType(db *DB, args database.Cmdline) resp.Reply {
	key := utils.BytesToString(args[0])
	entity, exist := db.GetEntity(key)
	if !exist {
		return reply.NewStatusReply("none")
	}
	switch entity.Data.(type) {
	case []byte:
		return reply.NewStatusReply("string")
	}
	//todo: 反射
	return reply.NewUnknownErrReply()
}

// ReName src dest
func execRename(db *DB, args database.Cmdline) resp.Reply {
	src := utils.BytesToString(args[0])
	dest := utils.BytesToString(args[1])
	entity, exist := db.GetEntity(src)
	if !exist {
		return reply.NewStatusReply("no such key")
	}
	db.PutEntity(dest, entity)
	db.Remove(src)
	return reply.NewOKReply()
}

// ReNameNx src dest
func execRenameNx(db *DB, args database.Cmdline) resp.Reply {
	src := utils.BytesToString(args[0])
	dest := utils.BytesToString(args[1])
	entity, exist := db.GetEntity(src)
	_, ok := db.GetEntity(dest)
	if ok {
		return reply.NewIntReply(0)
	}
	if !exist {
		return reply.NewStatusReply("no such key")
	}
	db.PutEntity(dest, entity)
	db.Remove(src)
	return reply.NewIntReply(1)
}

// keys *
func execKeys(db *DB, args database.Cmdline) resp.Reply {
	pattern := wildcard.CompilePattern(utils.BytesToString(args[0]))
	res := make([][]byte, 0)
	db.data.ForEach(func(key string, val any) bool {
		if pattern.IsMatch(key) {
			res = append(res, utils.StringToBytes(key))
		}
		return true
	})
	return reply.NewMultiBulkReply(res)
}
