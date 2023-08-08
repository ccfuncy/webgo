package cluster

import (
	"context"
	pool "github.com/jolestar/go-commons-pool/v2"
	database2 "redis/database"
	"redis/interface/database"
	"redis/interface/resp"
	"redis/interface/utils"
	"redis/lib/config"
	"redis/lib/consistenthash"
	"redis/lib/logger"
	"redis/resp/reply"
	"strings"
)

var router = makeRouter()

type ClusterDatabase struct {
	self string //自己名称

	nodes          []string                    //集群节点
	peerPicker     *consistenthash.NodeMap     //节点选择器
	peerConnection map[string]*pool.ObjectPool //节点连接池
	db             database.Database
}

func NewClusterDatabase() *ClusterDatabase {
	cluster := &ClusterDatabase{
		self:           config.Conf.Redis["self"].(string),
		db:             database2.NewStandaloneDataBase(),
		peerPicker:     consistenthash.NewNodeMap(nil),
		peerConnection: make(map[string]*pool.ObjectPool),
	}
	peers := config.Conf.Redis["peers"].(string)
	split := strings.Split(peers, ",")
	nodes := make([]string, 0, len(split)+1)
	for _, peer := range split {
		nodes = append(nodes, peer)
		objectPool := pool.NewObjectPoolWithDefaultConfig(context.Background(), &connectionFactory{
			peer,
		})
		cluster.peerConnection[peer] = objectPool

	}
	nodes = append(nodes, cluster.self)
	cluster.peerPicker.AddNodes(nodes...)
	cluster.nodes = nodes
	return cluster
}

func (c *ClusterDatabase) Exec(client resp.Connection, args database.Cmdline) (res resp.Reply) {
	defer func() {
		if err := recover(); err != nil {
			logger.Default().Error(err)
			res = reply.NewUnknownErrReply()
		}
	}()
	cmdName := strings.ToLower(utils.BytesToString(args[0]))
	cmdFunc, ok := router[cmdName]
	if !ok {
		return reply.NewArgNumErrReply("not support " + cmdName)
	}
	res = cmdFunc(c, client, args)
	return
}

func (c *ClusterDatabase) Close() {
	c.db.Close()
}

func (c *ClusterDatabase) AfterClose(client resp.Connection) {
	c.db.AfterClose(client)
}
