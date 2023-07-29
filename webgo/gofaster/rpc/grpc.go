package rpc

import (
	"google.golang.org/grpc"
	"net"
)

type FsGrpcServer struct {
	listen   net.Listener
	g        *grpc.Server
	register []func(*grpc.Server)
}

func NewFsGrpcServer(addr string, ops ...grpc.ServerOption) (*FsGrpcServer, error) {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	f := &FsGrpcServer{}
	f.listen = listen
	server := grpc.NewServer()
	f.g = server
	return f, nil
}

func (s *FsGrpcServer) Run() error {
	for _, f := range s.register {
		f(s.g)
	}
	return s.g.Serve(s.listen)
}
func (s *FsGrpcServer) Stop() {
	s.g.Stop()
}

func (s *FsGrpcServer) Register(f func(g *grpc.Server)) {
	s.register = append(s.register, f)
}
