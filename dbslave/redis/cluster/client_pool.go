package cluster

import (
	"context"
	"errors"
	pool "github.com/jolestar/go-commons-pool/v2"
	"redis/resp/client"
)

// 连接池工场
type connectionFactory struct {
	Peer string
}

func (c *connectionFactory) MakeObject(ctx context.Context) (*pool.PooledObject, error) {
	cl, err := client.NewClient(c.Peer)
	if err != nil {
		return nil, err
	}
	cl.Start()
	return pool.NewPooledObject(cl), nil
}

func (c *connectionFactory) DestroyObject(ctx context.Context, object *pool.PooledObject) error {
	cl, ok := object.Object.(*client.Client)
	if !ok {
		return errors.New("type mismatch")
	}
	cl.Close()
	return nil
}

func (c *connectionFactory) ValidateObject(ctx context.Context, object *pool.PooledObject) bool {
	return true
}

func (c *connectionFactory) ActivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}

func (c *connectionFactory) PassivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}
