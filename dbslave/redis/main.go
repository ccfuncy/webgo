package main

import (
	"redis/lib/config"
	"redis/lib/logger"
	"redis/resp/handler"
	"redis/tcp"
)

func main() {
	conf := config.Conf
	err := tcp.ListenAndServerWithSignal(&tcp.Config{
		Address: conf.Redis["self"].(string)}, handler.NewRespHandler())
	if err != nil {
		logger.Default().Error(err)
	}
}
