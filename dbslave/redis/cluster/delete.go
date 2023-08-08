package cluster

import (
	"redis/interface/database"
	"redis/interface/resp"
	"redis/resp/reply"
)

// del k1 k2 k3
func Delete(cluster *ClusterDatabase, conn resp.Connection, args database.Cmdline) resp.Reply {
	//todo: k1 k2 处于不同集群，需要广播删除指令，但是error这里处理不明
	replys := cluster.broadcast(conn, args)
	var res int64
	for _, re := range replys {
		if reply.IsErrReply(re) {
			return re
		}
		_, ok := re.(*reply.IntReply)
		if ok {
			res++
		}
	}
	return reply.NewIntReply(res)
}
