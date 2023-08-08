package cluster

import (
	"redis/interface/database"
	"redis/interface/resp"
)

// 无效转发
func Ping(cluster *ClusterDatabase, conn resp.Connection, args database.Cmdline) resp.Reply {
	exec := cluster.db.Exec(conn, args)
	return exec
}
