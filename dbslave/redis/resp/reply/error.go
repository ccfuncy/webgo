package reply

import (
	"fmt"
	"redis/interface/utils"
)

const (
	unknownErrBytes   = "-Err unknown\r\n"
	argNumErrBytes    = "-Err wrong number of arguments for %s command \r\n"
	syntaxErrBytes    = "-Err syntax error \r\n"
	wrongTypeErrBytes = "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
	protocolErrBytes  = "-Err Protocol error: %s \r\n"
)

type UnknownErrReply struct {
}

func NewUnknownErrReply() *UnknownErrReply {
	return &UnknownErrReply{}
}

func (u *UnknownErrReply) Error() string {
	return "Err unknown"
}

func (u *UnknownErrReply) ToBytes() []byte {
	return utils.StringToBytes(unknownErrBytes)
}

type ArgNumErrReply struct {
	cmd string
}

func NewArgNumErrReply(cmd string) *ArgNumErrReply {
	return &ArgNumErrReply{cmd: cmd}
}

func (a *ArgNumErrReply) Error() string {
	return fmt.Sprintf("Err wrong number of arguments for %s command", a.cmd)
}

func (a *ArgNumErrReply) ToBytes() []byte {
	return utils.StringToBytes(fmt.Sprintf(argNumErrBytes, a.cmd))
}

type SyntaxErrReply struct {
}

func (s *SyntaxErrReply) Error() string {
	return "Err syntax error"
}

func (s *SyntaxErrReply) ToBytes() []byte {
	return utils.StringToBytes(syntaxErrBytes)
}

func NewSyntaxErrReply() *SyntaxErrReply {
	return &SyntaxErrReply{}
}

type WrongTypeErrReply struct {
}

func (w *WrongTypeErrReply) Error() string {
	return "WRONGTYPE Operation against a key holding the wrong kind of value"
}

func (w *WrongTypeErrReply) ToBytes() []byte {
	return utils.StringToBytes(wrongTypeErrBytes)
}

func NewWrongTypeErrReply() *WrongTypeErrReply {
	return &WrongTypeErrReply{}
}

type ProtocolErrReply struct {
	Msg string
}

func (p *ProtocolErrReply) Error() string {
	return fmt.Sprintf("Err Protocol error: %s", p.Msg)
}

func (p *ProtocolErrReply) ToBytes() []byte {
	return utils.StringToBytes(fmt.Sprintf(protocolErrBytes, p.Msg))
}

func NewProtocolErrReply(msg string) *ProtocolErrReply {
	return &ProtocolErrReply{Msg: msg}
}
