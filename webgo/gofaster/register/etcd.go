package register

import (
	"context"
	"errors"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type FsEtcdRegister struct {
	cli *clientv3.Client
}

func (f *FsEtcdRegister) CreateCli(option Option) error {
	client, err := clientv3.New(clientv3.Config{Endpoints: option.Endpoints, DialTimeout: option.DialTimeout})
	if err != nil {
		return err
	}
	f.cli = client
	return nil
}

func (f *FsEtcdRegister) RegisterService(serviceName, host string, port int) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
	defer cancelFunc()
	_, err := f.cli.Put(ctx, serviceName, fmt.Sprintf("%s:%d", host, port))
	return err
}

func (f *FsEtcdRegister) GetValue(serviceName string) (string, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
	defer cancelFunc()
	instance, err := f.cli.Get(ctx, serviceName)
	if err != nil {
		return "", err
	}
	kvs := instance.Kvs
	if len(kvs) == 0 {
		return "", errors.New("no value")
	}
	return kvs[0].String(), nil
}

func (f *FsEtcdRegister) Close() error {
	return f.cli.Close()
}
