package reply

import "redis/interface/utils"

const (
	pongBytes      = "+PONG\r\n"
	okBytes        = "+OK\r\n"
	nullBulkBytes  = "$-1\r\n" //nil
	emptyBulkBytes = "*0\r\n"
	noBytes        = ""
)

type PongReply struct{}

func NewPongReply() *PongReply {
	return &PongReply{}
}

func (p *PongReply) ToBytes() []byte {
	return utils.StringToBytes(pongBytes)
}

type OKReply struct{}

func (O *OKReply) ToBytes() []byte {
	return utils.StringToBytes(okBytes)
}

func NewOKReply() *OKReply {
	return &OKReply{}
}

type NullBulkReply struct {
}

func (N *NullBulkReply) ToBytes() []byte {
	return utils.StringToBytes(nullBulkBytes)
}

func NewNULLBulkReply() *NullBulkReply {
	return &NullBulkReply{}
}

type EmptyBulkReply struct {
}

func (e *EmptyBulkReply) ToBytes() []byte {
	return utils.StringToBytes(emptyBulkBytes)
}

func NewEmptyBulkReply() *EmptyBulkReply {
	return &EmptyBulkReply{}
}

type NoReply struct {
}

func (n *NoReply) ToBytes() []byte {
	return utils.StringToBytes(noBytes)
}

func NewNoReply() *NoReply {
	return &NoReply{}
}
