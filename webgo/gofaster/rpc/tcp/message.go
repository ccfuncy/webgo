package tcp

type MessageType byte

const (
	msgRequest MessageType = iota
	msgResponse
	msgPing
	msgPong
)
