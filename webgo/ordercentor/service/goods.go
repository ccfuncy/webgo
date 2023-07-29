package service

import "gofaster/rpc"

type GoodsServices struct {
	Find func(map[string]any) ([]byte, error) `fsrpc:"GET,/goods/find"`
}

func (g GoodsServices) Env() *rpc.HttpConfig {
	return &rpc.HttpConfig{
		Protocol: "http",
		Host:     "localhost",
		Port:     9002,
	}
}
