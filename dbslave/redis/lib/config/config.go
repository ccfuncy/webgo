package config

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"redis/lib/logger"
)

var Conf = &Config{
	logger: logger.Default(),
	Redis:  make(map[string]any),
}

type Config struct {
	logger *logger.Logger
	Redis  map[string]any
}

func init() {
	loadToml()
}

func loadToml() {
	configFile := flag.String("conf", "conf/app.toml", "app config file")
	flag.Parse()
	if _, err := os.Stat(*configFile); err != nil {
		Conf.logger.Info(fmt.Sprintf("%s file not load.because not exist!", *configFile))
		return
	}
	_, err := toml.DecodeFile(*configFile, Conf)
	if err != nil {
		Conf.logger.Info(fmt.Sprintf("%s decode fail. please check format ! %s", *configFile, err.Error()))
		return
	}
}
