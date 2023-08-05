package register

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

type FsNacosRegister struct {
	cli naming_client.INamingClient
}

func (f *FsNacosRegister) CreateCli(option Option) error {

	//clientConfig := *constant.NewClientConfig(
	//	constant.WithNamespaceId(""),
	//	constant.WithTimeoutMs(5000),
	//	constant.WithNotLoadCacheAtStart(true),
	//	constant.WithLogDir("/tmp/nacos/logger"),
	//	constant.WithCacheDir("/tmp/nacos/cache"),
	//	constant.WithLogLevel("debug"),
	//)
	//serverConfigs := []constant.ServerConfig{*constant.NewServerConfig(
	//	"127.0.0.1",
	//	8848,
	//	constant.WithScheme("http"),
	//	constant.WithContextPath("/nacos"),
	//)}
	client, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  option.NacosClientConfig,
		ServerConfigs: option.NacosServerConfigs,
	})
	if err != nil {
		return err
	}
	f.cli = client
	return nil
}

func (f *FsNacosRegister) RegisterService(serviceName, host string, port int) error {
	_, err := f.cli.RegisterInstance(vo.RegisterInstanceParam{
		Ip:       host,
		Port:     uint64(port),
		Weight:   10,
		Enable:   true,
		Healthy:  true,
		Metadata: map[string]string{"idc": "shanghai"},
		//ClusterName: "",
		ServiceName: serviceName,
		//GroupName: "",
		Ephemeral: true,
	})
	return err
}

func (f *FsNacosRegister) GetValue(serviceName string) (string, error) {
	instance, err := f.cli.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: serviceName,
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", instance.Ip, instance.Port), nil
}

func (f *FsNacosRegister) Close() error {
	return nil
}
