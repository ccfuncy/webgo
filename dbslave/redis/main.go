package main

import (
	"fmt"
	"redis/lib/config"
	"redis/lib/logger"
	"redis/resp/handler"
	"redis/tcp"
)

func main() {
	conf := config.Conf
	err := tcp.ListenAndServerWithSignal(&tcp.Config{
		Address: fmt.Sprintf("%s:%d", conf.Redis["bind"], conf.Redis["port"])}, handler.NewRespHandler())
	if err != nil {
		logger.Default().Error(err)
	}
}
