package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"gofaster"
	"gofaster/log"
	"gofaster/rpc"
	"gofaster/rpc/tcp"
	"goodscentor/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"ordercentor/api"
	"ordercentor/service"
)

func main() {
	engine := gofaster.Default()
	group := engine.Group("order")
	group.Use(gofaster.Logging)
	client := rpc.NewFsHttpClient()
	client.RegisterHttpServiceName("goods", &service.GoodsServices{})
	group.Get("/find", func(ctx *gofaster.Context) {
		//通过商品中心查询调用
		params := make(map[string]any)
		params["id"] = 12
		body, err := client.Do("goods", "Find").(*service.GoodsServices).Find(params)
		//get, err := client.Get("http://localhost:9002/goods/find", params)
		if err != nil {
			panic(err)
		}
		v := model.Result{}
		err = json.Unmarshal(body, &v)
		if err != nil {
			panic(err)
		}
		ctx.JSON(http.StatusOK, v)
	})
	group.Get("/findGrpc", func(ctx *gofaster.Context) {
		conn, err := grpc.Dial("localhost:9111", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			panic(err)
		}
		defer conn.Close()
		apiClient := api.NewGoodsApiClient(conn)
		response, err := apiClient.Find(context.Background(), &api.GoodsRequest{})
		if err != nil {
			panic(err)
		}
		ctx.JSON(http.StatusOK, response)
	})
	group.Get("/findTcp", func(ctx *gofaster.Context) {
		proxy := tcp.NewFsTcpClientProxy(tcp.DefaultOption)
		params := make([]any, 1)
		params[0] = int64(1)
		gob.Register(&model.Goods{})
		gob.Register(&model.Result{})
		res, err := proxy.Call(context.Background(), "goods", "Find", params)
		if err != nil {
			log.Default().Error(err)
		}
		ctx.JSON(http.StatusOK, res)
	})
	engine.Run(":9003")

}
