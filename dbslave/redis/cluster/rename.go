package cluster

import (
	"redis/interface/database"
	"redis/interface/resp"
	"redis/interface/utils"
	"redis/resp/reply"
)

// 双操作数
// rename k1 k2
func Rename(cluster *ClusterDatabase, conn resp.Connection, args database.Cmdline) resp.Reply {
	if len(args) != 3 {
		return reply.NewStandardErrReply("Err wrong number args")
	}
	src := utils.BytesToString(args[1])
	dest := utils.BytesToString(args[2])
	srcNode := cluster.peerPicker.PickNode(src)
	destNode := cluster.peerPicker.PickNode(dest)
	if srcNode != destNode {
		return reply.NewStandardErrReply("rename args src node must equal dest node")
	}
	return cluster.relay(srcNode, conn, args)
}
