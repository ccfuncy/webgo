package tcp

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"gofaster/log"
	"net"
	"sync/atomic"
	"time"
)

type FsRpcClient interface {
	Connect() error
	Invoke(ctx context.Context, serviceName, methodName string, args []any) (any, error)
	Close() error
}

type TcpClientOption struct {
	Retries           int
	ConnectionTimeout time.Duration
	SerializeType     SerializeType
	CompressType      CompressType
	Host              string
	Port              int
}

var DefaultOption = TcpClientOption{
	Retries:           3,
	ConnectionTimeout: 5 * time.Second,
	SerializeType:     Gob,
	CompressType:      Gzip,
	Host:              "localhost",
	Port:              9111,
}

type FsTcpClient struct {
	conn   net.Conn
	option TcpClientOption
}

func NewFsTcpClient(option TcpClientOption) *FsTcpClient {
	return &FsTcpClient{option: option}
}

func (f *FsTcpClient) Connect() error {
	addr := fmt.Sprintf("%s:%d", f.option.Host, f.option.Port)
	conn, err := net.DialTimeout("tcp", addr, f.option.ConnectionTimeout)
	if err != nil {
		return err
	}
	f.conn = conn
	return nil
}

var reqId int64

func (f *FsTcpClient) Invoke(ctx context.Context, serviceName, methodName string, args []any) (any, error) {
	//包装request对象，在发送
	req := &FsRpcRequest{
		RequestId:   atomic.AddInt64(&reqId, 1),
		ServiceName: serviceName,
		MethodName:  methodName,
		Args:        args,
	}
	//header
	header := make([]byte, 17)
	header[0] = MagicNumber
	header[1] = Version
	header[6] = byte(MessageType(msgRequest))
	header[7] = byte(f.option.CompressType)
	header[8] = byte(f.option.SerializeType)
	binary.BigEndian.PutUint64(header[9:], uint64(req.RequestId))
	//先序列化，在压缩
	serializer := loadSerializer(f.option.SerializeType)
	bytes, err := serializer.Serialize(req)
	if err != nil {
		return nil, err
	}
	compress := loadCompress(f.option.CompressType)
	body, err := compress.Compress(bytes)
	if err != nil {
		return nil, err
	}
	binary.BigEndian.PutUint32(header[2:6], uint32(len(body)+17))

	_, err = f.conn.Write(header[:])
	if err != nil {
		return nil, err
	}
	_, err = f.conn.Write(body[:])
	if err != nil {
		return nil, err
	}
	rspChan := make(chan *FsRpcResponse)
	go f.readHandler(rspChan)
	rsp := <-rspChan
	return rsp, nil
}

func (f *FsTcpClient) Close() error {
	if f.conn != nil {
		return f.conn.Close()
	}
	return nil
}

func (f *FsTcpClient) readHandler(rspChan chan *FsRpcResponse) {
	defer func() {
		if err := recover(); err != nil {
			log.Default().Error(err)
			f.conn.Close()
		}
	}()
	for {
		message, err := decodeFrame(f.conn)
		if err != nil {
			log.Default().Error(err)
			rsp := &FsRpcResponse{
				Code: 500,
				Msg:  err.Error(),
			}
			rspChan <- rsp
			return
		}
		if message.Header.MessageType == msgResponse {
			response := message.Data.(*FsRpcResponse)
			rspChan <- response
			return
		}
	}
}

type FsTcpClientProxy struct {
	client *FsTcpClient
	option TcpClientOption
}

func (p *FsTcpClientProxy) Call(ctx context.Context, serviceName, methodName string, args []any) (any, error) {
	client := NewFsTcpClient(p.option)
	p.client = client
	err := client.Connect()
	if err != nil {
		return nil, err
	}
	for i := 0; i < p.option.Retries; i++ {
		res, err := client.Invoke(ctx, serviceName, methodName, args)
		if err != nil {
			if i >= p.option.Retries-1 {
				log.Default().Error("already retry all time")
				client.Close()
				return nil, err
			}
			continue
		}
		client.Close()
		return res, nil
	}
	return nil, errors.New("retry time is 0")
}

func NewFsTcpClientProxy(option TcpClientOption) *FsTcpClientProxy {
	return &FsTcpClientProxy{option: option}
}
