package resp

import "net"

type Connection interface {
	Write([]byte) error
	GetDBIndex() int
	SelectDB(int)
	RemoteAddr() net.Addr
}
