package cluster

import (
	"context"
	"errors"
	"redis/interface/database"
	"redis/interface/resp"
	"redis/interface/utils"
	"redis/resp/client"
	"redis/resp/reply"
	"strconv"
)

func (c *ClusterDatabase) getPeerClient(peer string) (*client.Client, error) {
	pool, ok := c.peerConnection[peer]
	if !ok {
		return nil, errors.New("connection not found")
	}
	object, err := pool.BorrowObject(context.Background())
	if err != nil {
		return nil, err
	}
	cl, ok := object.(*client.Client)
	if !ok {
		return nil, errors.New("wrong type")
	}
	return cl, nil
}

func (c *ClusterDatabase) returnPeerClient(peer string, cl *client.Client) error {
	pool, ok := c.peerConnection[peer]
	if !ok {
		return errors.New("connection not found")
	}
	return pool.ReturnObject(context.Background(), cl)
}

// 转发
func (c *ClusterDatabase) relay(peer string, conn resp.Connection, args database.Cmdline) resp.Reply {
	if peer == c.self {
		return c.db.Exec(conn, args)
	}
	peerClient, err := c.getPeerClient(peer)
	if err != nil {
		return reply.NewStandardErrReply(err.Error())
	}
	defer func() {
		_ = c.returnPeerClient(peer, peerClient)
	}()
	//选择库,
	peerClient.Send(utils.ToCmdLine("select", strconv.Itoa(conn.GetDBIndex())))
	return peerClient.Send(args)
}

// 广播
func (c *ClusterDatabase) broadcast(connection resp.Connection, args database.Cmdline) map[string]resp.Reply {
	m := make(map[string]resp.Reply)
	for _, node := range c.nodes {
		relay := c.relay(node, connection, args)
		m[node] = relay
	}
	return m
}
