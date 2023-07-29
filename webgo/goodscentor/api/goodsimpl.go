package api

import (
	"context"
)

type GoodsRpcService struct {
}

func (GoodsRpcService) Find(context.Context, *GoodsRequest) (*GoodsResponse, error) {
	goods := &Goods{Id: 1000, Name: "9002商品"}
	return &GoodsResponse{
		Code: 200,
		Msg:  "success",
		Data: goods,
	}, nil
}
func (GoodsRpcService) mustEmbedUnimplementedGoodsApiServer() {}
