package cluster

import (
	"redis/interface/database"
	"redis/interface/resp"
)

func Select(cluster *ClusterDatabase, conn resp.Connection, args database.Cmdline) resp.Reply {
	//todo: 其实这里是将dbindex放入conn种
	return cluster.db.Exec(conn, args)
}
