package cluster

import (
	"redis/interface/database"
	"redis/interface/resp"
	"redis/interface/utils"
)

type ClusterCmdFunc func(cluster *ClusterDatabase, conn resp.Connection, args database.Cmdline) resp.Reply

func makeRouter() map[string]ClusterCmdFunc {
	m := make(map[string]ClusterCmdFunc)
	//todo:exist 存疑
	m["exists"] = defaultFunc //exists k1
	m["type"] = defaultFunc
	m["set"] = defaultFunc
	m["setnx"] = defaultFunc
	m["getset"] = defaultFunc
	m["get"] = defaultFunc
	m["Ping"] = Ping
	// nx 代表不覆盖，会优先查询key是否存在
	m["rename"] = Rename
	m["renamenx"] = Rename

	m["flushdb"] = FlushDB
	m["del"] = Delete
	return m
}

// Get k set k v
// 单个操作数走默认
func defaultFunc(cluster *ClusterDatabase, conn resp.Connection, args database.Cmdline) resp.Reply {
	key := utils.BytesToString(args[1])
	node := cluster.peerPicker.PickNode(key)
	return cluster.relay(node, conn, args)
}
