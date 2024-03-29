// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package api

import (
	context "context"
	grpc "google.golang.org/grpc"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// GoodsApiClient is the client API for GoodsApi service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GoodsApiClient interface {
	Find(ctx context.Context, in *GoodsRequest, opts ...grpc.CallOption) (*GoodsResponse, error)
}

type goodsApiClient struct {
	cc grpc.ClientConnInterface
}

func NewGoodsApiClient(cc grpc.ClientConnInterface) GoodsApiClient {
	return &goodsApiClient{cc}
}

func (c *goodsApiClient) Find(ctx context.Context, in *GoodsRequest, opts ...grpc.CallOption) (*GoodsResponse, error) {
	out := new(GoodsResponse)
	err := c.cc.Invoke(ctx, "/api.GoodsApi/Find", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GoodsApiServer is the server API for GoodsApi service.
// All implementations must embed UnimplementedGoodsApiServer
// for forward compatibility
type GoodsApiServer interface {
	Find(context.Context, *GoodsRequest) (*GoodsResponse, error)
	mustEmbedUnimplementedGoodsApiServer()
}

// UnsafeGoodsApiServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GoodsApiServer will
// result in compilation errors.
type UnsafeGoodsApiServer interface {
	mustEmbedUnimplementedGoodsApiServer()
}

func RegisterGoodsApiServer(s grpc.ServiceRegistrar, srv GoodsApiServer) {
	s.RegisterService(&GoodsApi_ServiceDesc, srv)
}

func _GoodsApi_Find_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GoodsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GoodsApiServer).Find(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.GoodsApi/Find",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GoodsApiServer).Find(ctx, req.(*GoodsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// GoodsApi_ServiceDesc is the grpc.ServiceDesc for GoodsApi service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GoodsApi_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.GoodsApi",
	HandlerType: (*GoodsApiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Find",
			Handler:    _GoodsApi_Find_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/goods.proto",
}
