package tcp

import (
	"encoding/binary"
	"gofaster/log"
	"net"
	"reflect"
)

type FsRpcServer interface {
	Register(name string, service any)
	Run()
	Stop()
}
type FsTcpServer struct {
	listen     net.Listener
	serviceMap map[string]any
}

func NewFsTcpServer(addr string) (*FsTcpServer, error) {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &FsTcpServer{listen: listen, serviceMap: make(map[string]any)}, nil
}

func (f *FsTcpServer) Register(name string, service any) {
	t := reflect.TypeOf(service)
	if t.Kind() != reflect.Pointer {
		panic("service must be a pointer")
	}
	f.serviceMap[name] = service
}

type FsTcpConn struct {
	conn    net.Conn
	rspChan chan *FsRpcResponse
}

func (c FsTcpConn) Send(rsp *FsRpcResponse) error {
	if rsp.Code != 200 {
		//默认数据发送
	}
	//header
	header := make([]byte, 17)
	header[0] = MagicNumber
	header[1] = Version
	header[6] = byte(MessageType(msgResponse))
	header[7] = byte(rsp.CompressType)
	header[8] = byte(rsp.SerializeType)
	binary.BigEndian.PutUint64(header[9:], uint64(rsp.RequestId))
	//先序列化，在压缩
	serializer := loadSerializer(rsp.SerializeType)
	bytes, err := serializer.Serialize(rsp.Data)
	if err != nil {
		return err
	}
	compress := loadCompress(rsp.CompressType)
	body, err := compress.Compress(bytes)
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint32(header[2:6], uint32(len(body)+17))
	_, err = c.conn.Write(header[:])
	if err != nil {
		return err
	}
	_, err = c.conn.Write(body[:])
	if err != nil {
		return err
	}
	return nil
}

func (f *FsTcpServer) Run() {
	for {
		conn, err := f.listen.Accept()
		if err != nil {
			log.Default().Error(err)
			continue
		}
		tcpConn := &FsTcpConn{
			conn:    conn,
			rspChan: make(chan *FsRpcResponse),
		}
		//接收数据，解码，请求业务获取结果，发送到response
		//获取结果，编码，发送
		go f.readHandler(tcpConn)
		go f.writeHandler(tcpConn)
	}
}

func (f *FsTcpServer) Stop() {
	_ = f.listen.Close()
}

func (f *FsTcpServer) readHandler(conn *FsTcpConn) {
	defer func() {
		if err := recover(); err != nil {
			log.Default().Error(err)
			conn.conn.Close()
		}
	}()
	//接收数据
	//解码
	message, err := decodeFrame(conn.conn)
	if err != nil {
		rsp := &FsRpcResponse{
			Code: 500,
			Msg:  err.Error(),
		}
		conn.rspChan <- rsp
		return
	}
	if message.Header.MessageType == msgRequest {
		request := message.Data.(*FsRpcRequest)
		rsp := &FsRpcResponse{
			RequestId:     request.RequestId,
			CompressType:  message.Header.CompressType,
			SerializeType: message.Header.SerializeType,
		}
		service, ok := f.serviceMap[request.ServiceName]
		if !ok {
			rsp := &FsRpcResponse{
				Code: 500,
				Msg:  "no service name found",
			}
			conn.rspChan <- rsp
			return
		}
		method := reflect.ValueOf(service).MethodByName(request.MethodName)
		if method.IsNil() {
			rsp := &FsRpcResponse{
				Code: 500,
				Msg:  "no method name found",
			}
			conn.rspChan <- rsp
			return
		}
		//调用方法
		args := make([]reflect.Value, len(request.Args))
		for i, arg := range request.Args {
			args[i] = reflect.ValueOf(arg)
		}
		//args[0] = reflect.ValueOf(service)

		res := method.Call(args)
		results := make([]any, len(res))
		for i, re := range res {
			results[i] = re.Interface()
		}
		err, ok := results[len(results)-1].(error)
		if ok {
			rsp := &FsRpcResponse{
				Code: 500,
				Msg:  err.Error(),
			}
			conn.rspChan <- rsp
			return
		}
		rsp.Code = 200
		rsp.Msg = "success"
		rsp.Data = results[0]
		conn.rspChan <- rsp
	}
}

func (f *FsTcpServer) writeHandler(conn *FsTcpConn) {
	select {
	case rsp := <-conn.rspChan:
		err := conn.Send(rsp)
		defer conn.conn.Close()
		if err != nil {
			log.Default().Error(err)
		}
	}
}

func loadSerializer(serializeType SerializeType) Serializer {
	switch serializeType {
	case Gob:
		return GobSerializer{}
	}
	return nil
}

func loadCompress(compressType CompressType) CompressInterface {
	switch compressType {
	case Gzip:
		return GzipCompress{}
	}
	return nil
}
