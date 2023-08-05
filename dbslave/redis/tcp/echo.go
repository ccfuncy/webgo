package tcp

import (
	"bufio"
	"context"
	"io"
	"net"
	"redis/lib/logger"
	"redis/lib/sync/atomic"
	"redis/lib/sync/wait"
	"sync"
	"time"
)

type EchoClient struct {
	// 防止工作还没做完就被关闭
	wait wait.Wait

	conn net.Conn
}

func (e *EchoClient) Close() error {
	e.wait.WaitWithTimeout(time.Second * 10)
	_ = e.conn.Close()
	return nil
}

type EchoHandler struct {
	activeConn sync.Map
	closing    atomic.Boolean
}

func NewEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

func (e *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	if e.closing.Get() {
		_ = conn.Close()
		return
	}
	client := &EchoClient{
		conn: conn,
	}
	e.activeConn.Store(client, struct{}{})
	reader := bufio.NewReader(conn)
	for true {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.Default().Info("connection close")
				e.activeConn.Delete(client)
			} else {
				logger.Default().Error(err)
			}
			return
		}
		client.wait.Add(1)
		_, _ = conn.Write([]byte(msg))
		client.wait.Done()
	}
}

func (e *EchoHandler) Close() error {
	logger.Default().Info("handler shutting down ")
	e.closing.Set(true)
	e.activeConn.Range(func(key, value any) bool {
		client := key.(*EchoClient)
		_ = client.Close()
		return true
	})
	return nil
}
