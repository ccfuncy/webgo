package cluster

import (
	"redis/interface/database"
	"redis/interface/resp"
	"redis/interface/utils"
	"redis/resp/reply"
)

func FlushDB(cluster *ClusterDatabase, conn resp.Connection, args database.Cmdline) resp.Reply {
	if len(args) != 1 {
		return reply.NewArgNumErrReply(utils.BytesToString(args[0]))
	}
	broadcast := cluster.broadcast(conn, args)
	for _, r := range broadcast {
		if reply.IsErrReply(r) {
			return r
		}
	}
	return reply.NewOKReply()
}
