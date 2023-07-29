package tcp

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
)

const (
	MagicNumber byte = 0x1d
	Version          = 0x01
)

type Header struct {
	MagicNumber   byte          //魔法数 1字节
	Version       byte          //版本号 1字节
	FullLength    int32         //消息长度 4字节
	MessageType   MessageType   //消息类型 1字节
	SerializeType SerializeType //序列化类型 1字节
	CompressType  CompressType  //压缩类型 1字节
	RequestId     int64         //请求ID 8字节
}
type FsRpcMessage struct {
	Header Header
	Data   any
}

type FsRpcRequest struct {
	RequestId   int64
	ServiceName string
	MethodName  string
	Args        []any
}

type FsRpcResponse struct {
	RequestId     int64
	Code          int16
	Msg           string
	CompressType  CompressType
	SerializeType SerializeType
	Data          any
}

func decodeFrame(conn net.Conn) (*FsRpcMessage, error) {
	//header 17字节
	header := make([]byte, 17)
	_, err := io.ReadFull(conn, header)
	if err != nil {
		return nil, err
	}
	if header[0] != MagicNumber {
		return nil, errors.New("magic number error")
	}
	//header
	msg := &FsRpcMessage{
		Header: Header{
			MagicNumber:   header[0],
			Version:       header[1],
			FullLength:    int32(binary.BigEndian.Uint32(header[2:6])),
			MessageType:   MessageType(header[6]),
			SerializeType: SerializeType(header[7]),
			CompressType:  CompressType(header[8]),
			RequestId:     int64(binary.BigEndian.Uint64(header[9:])),
		},
	}
	//body
	bodyLength := msg.Header.FullLength - 17
	body := make([]byte, bodyLength)
	_, err = io.ReadFull(conn, body)
	if err != nil {
		return nil, err
	}
	//编码时序列化在压缩，解码时需解压再反序列化
	compress := loadCompress(msg.Header.CompressType)
	if compress == nil {
		return nil, errors.New("no compress interface")
	}
	bytes, err := compress.UnCompress(body)
	if err != nil {
		return nil, err
	}
	serializer := loadSerializer(msg.Header.SerializeType)
	if serializer == nil {
		return nil, errors.New("no serializer")
	}
	var rsp any
	switch msg.Header.MessageType {
	case msgRequest:
		rsp = &FsRpcRequest{}
	case msgResponse:
		rsp = &FsRpcResponse{}
	}
	err = serializer.Deserialize(bytes, rsp)
	if err != nil {
		return nil, err
	}
	msg.Data = rsp
	return msg, nil
}
