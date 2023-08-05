package connection

import (
	"net"
	"redis/lib/sync/wait"
	"sync"
	"time"
)

// Connection 协议层对每个客户端的描述
type Connection struct {
	conn         net.Conn
	waitingReply wait.Wait  //保证在关闭时处理完毕
	mu           sync.Mutex //保证并发
	selectDB     int        //现在的用户在操作哪个DB
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{conn: conn}
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Connection) Close() error {
	c.waitingReply.WaitWithTimeout(time.Second * 10)
	_ = c.conn.Close()
	return nil
}

func (c *Connection) Write(bytes []byte) error {
	if len(bytes) == 0 {
		return nil
	}
	//需要加锁，如果同时两个协程写数据
	c.mu.Lock()
	c.waitingReply.Add(1)
	defer func() {
		c.mu.Unlock()
		c.waitingReply.Done()
	}()
	_, err := c.conn.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}

func (c *Connection) GetDBIndex() int {
	return c.selectDB
}

func (c *Connection) SelectDB(i int) {
	c.selectDB = i
}
