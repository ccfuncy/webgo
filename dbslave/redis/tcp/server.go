package tcp

import (
	"context"
	"net"
	"os"
	"os/signal"
	"redis/interface/tcp"
	"redis/lib/logger"
	"redis/lib/sync/wait"
	"syscall"
)

type Config struct {
	// 监听地址字符串
	Address string
}

func ListenAndServerWithSignal(config *Config, handler tcp.Handler) error {
	listen, err := net.Listen("tcp", config.Address)
	if err != nil {
		logger.Default().Error(err)
		return err
	}
	logger.Default().Info("start listen" + config.Address)
	// 用于关闭链接通信
	closeChan := make(chan struct{})
	// 用于监听系统退出信号，通过notify 注册上去
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()

	ListenAndServe(listen, handler, closeChan)
	return nil
}
func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	defer func() {
		_ = listener.Close()
		_ = handler.Close()
	}()
	go func() {
		<-closeChan
		logger.Default().Info("shutting down")
		_ = listener.Close()
		_ = handler.Close()
	}()
	ctx := context.Background()
	var waitDone wait.Wait
	for true {
		conn, err := listener.Accept()
		if err != nil {
			logger.Default().Error(err)
			break
		}
		logger.Default().Info("accepted link " + conn.RemoteAddr().String())
		waitDone.Add(1)
		go func() {
			defer func() {
				waitDone.Done()
			}()
			handler.Handle(ctx, conn)
		}()
	}
	// 防止accept error 后还有未执行完毕的协程
	waitDone.Wait()
}
