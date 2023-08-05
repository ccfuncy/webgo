package reply

import (
	"bytes"
	"fmt"
	"redis/interface/resp"
	"redis/interface/utils"
)

const (
	nullBulkReplyBytes = "$-1"
	CRLF               = "\r\n"
)

type BulkReply struct {
	Arg []byte
}

func (b *BulkReply) ToBytes() []byte {
	var buf bytes.Buffer
	if len(b.Arg) == 0 {
		buf.WriteString(nullBulkBytes + CRLF)
	} else {
		buf.WriteString(fmt.Sprintf("$%d%s%s%s", len(b.Arg), CRLF, b.Arg, CRLF))
	}
	return buf.Bytes()
}

func NewBulkReply(arg []byte) *BulkReply {
	return &BulkReply{arg}
}

type MultiBulkReply struct {
	Args [][]byte
}

func (m *MultiBulkReply) ToBytes() []byte {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("*%d%s", len(m.Args), CRLF))
	for _, arg := range m.Args {
		if arg == nil {
			buf.WriteString(nullBulkBytes + CRLF)
		} else {
			buf.WriteString(fmt.Sprintf("$%d%s%s%s", len(arg), CRLF, arg, CRLF))
		}
	}
	return buf.Bytes()
}

func NewMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{args}
}

type StatusReply struct {
	Status string
}

func (s *StatusReply) ToBytes() []byte {
	return utils.StringToBytes(fmt.Sprintf("+%s%s", s.Status, CRLF))
}

func NewStatusReply(status string) *StatusReply {
	return &StatusReply{status}
}

type IntReply struct {
	Code int64
}

func (i *IntReply) ToBytes() []byte {
	return utils.StringToBytes(fmt.Sprintf(":%d%s", i.Code, CRLF))
}

func NewIntReply(code int64) *IntReply {
	return &IntReply{Code: code}
}

type StandardErrReply struct {
	Status string
}

func (s *StandardErrReply) Error() string {
	return s.Status
}

func (s *StandardErrReply) ToBytes() []byte {
	return utils.StringToBytes(fmt.Sprintf("-%s%s", s.Status, CRLF))
}

func NewStandardErrReply(status string) *StandardErrReply {
	return &StandardErrReply{status}
}

func IsErrReply(reply resp.Reply) bool {
	return reply.ToBytes()[0] == '-'
}
