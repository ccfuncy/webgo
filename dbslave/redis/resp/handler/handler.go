package handler

import (
	"context"
	"io"
	"net"
	"redis/database"
	databaseface "redis/interface/database"
	"redis/interface/utils"
	"redis/lib/logger"
	"redis/lib/sync/atomic"
	"redis/resp/connection"
	"redis/resp/parser"
	"redis/resp/reply"
	"strings"
	"sync"
)

var unknownErrReplyBytes []byte = utils.StringToBytes("-ERR unknow\r\n")

type RespHandler struct {
	activeConn sync.Map
	closing    atomic.Boolean
	db         databaseface.Database
}

func NewRespHandler() *RespHandler {
	database := database.NewDataBase()
	return &RespHandler{db: database}
}

func (r *RespHandler) closeClient(client *connection.Connection) {
	//关闭的单个链接
	_ = client.Close()
	r.db.AfterClose(client)
	r.activeConn.Delete(client)
}

func (r *RespHandler) Handle(ctx context.Context, conn net.Conn) {
	if r.closing.Get() {
		_ = conn.Close()
		return
	}
	client := connection.NewConnection(conn)
	r.activeConn.Store(client, struct{}{})
	ch := parser.ParseStream(conn)
	for payload := range ch {
		//err
		if payload.Err != nil {
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				//客户端挥手
				r.closeClient(client)
				logger.Default().Info("connection close: " + client.RemoteAddr().String())
				return
			}
			//protocol err
			errReply := reply.NewProtocolErrReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes())
			if err != nil {
				r.closeClient(client)
				logger.Default().Info("connection close: " + client.RemoteAddr().String())
				return
			}
			continue
		}
		//exec
		if payload.Data == nil {
			logger.Default().Info("connection(" + client.RemoteAddr().String() + ") send  empty data")
			continue
		}
		//多行字符串才是指令
		multiBulkReply, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Default().Info("connection(" + client.RemoteAddr().String() + ") send :" +
				utils.BytesToString(payload.Data.ToBytes()))
			continue
		}
		exec := r.db.Exec(client, multiBulkReply.Args)
		if exec != nil {
			client.Write(exec.ToBytes())
		} else {
			client.Write(unknownErrReplyBytes)
		}

	}
}

func (r *RespHandler) Close() error {
	logger.Default().Info("handler shutting down")
	r.closing.Set(true)
	r.activeConn.Range(func(key, value any) bool {
		c := key.(*connection.Connection)
		_ = c.Close()
		return true
	})
	r.db.Close()
	return nil
}
