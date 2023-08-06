package connection

import "net"

type FakeConnection struct {
	index int
	addr  *FakeAddr
}

func NewFakeConnection() *FakeConnection {
	return &FakeConnection{addr: &FakeAddr{}}
}

func (f *FakeConnection) Write(bytes []byte) error {
	//TODO implement me
	panic("implement me")
}

func (f *FakeConnection) GetDBIndex() int {
	return f.index
}

func (f *FakeConnection) SelectDB(i int) {
	f.index = i
}

func (f *FakeConnection) RemoteAddr() net.Addr {
	return f.addr
}

type FakeAddr struct {
}

func (f FakeAddr) Network() string {
	//TODO implement me
	panic("implement me")
}

func (f FakeAddr) String() string {
	return "aof.handler"
}
