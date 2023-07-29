package register

import (
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"time"
)

type Option struct {
	Endpoints          []string
	DialTimeout        time.Duration
	ServiceName        string
	Host               string
	Port               int
	NacosServerConfigs []constant.ServerConfig
	NacosClientConfig  *constant.ClientConfig
}

type FsRegister interface {
	CreateCli(option Option) error
	RegisterService(serviceName, host string, port int) error
	GetValue(serviceName string) (string, error)
	Close() error
}
