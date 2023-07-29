package main

import (
	"encoding/gob"
	"gofaster"
	"gofaster/log"
	"gofaster/rpc/tcp"
	"goodscentor/model"
	"goodscentor/service"
	"net/http"
)

func main() {
	engine := gofaster.Default()
	group := engine.Group("goods")
	group.Use(gofaster.Logging)
	group.Get("/find", func(ctx *gofaster.Context) {
		goods := &model.Goods{Id: 1000, Name: "9002商品"}
		ctx.JSON(http.StatusOK, &model.Result{
			Code: 200,
			Msg:  "success",
			Data: goods,
		})
	})
	//server, _ := rpc.NewFsGrpcServer(":9111")
	//server.Register(func(g *grpc.Server) {
	//	api.RegisterGoodsApiServer(g, &api.GoodsRpcService{})
	//})
	//err := server.Run()
	//if err != nil {
	//	panic(err)
	//}
	server, err := tcp.NewFsTcpServer(":9111")
	if err != nil {
		log.Default().Error(err)
	}
	gob.Register(&model.Goods{})
	gob.Register(&model.Result{})
	server.Register("goods", &service.GoodsService{})
	server.Run()
	engine.Run(":9002")

}
