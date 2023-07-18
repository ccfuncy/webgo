package config

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	fslog "gofaster/log"
	"os"
)

var Conf = &FsConfig{
	logger: fslog.Default(),
}

type FsConfig struct {
	logger   *fslog.Logger
	Log      map[string]any
	Pool     map[string]any
	Template map[string]any
}

func init() {
	loadToml()
}

func loadToml() {
	configFile := flag.String("Conf", "Conf/app.toml", "app config file")
	flag.Parse()
	if _, err := os.Stat(*configFile); err != nil {
		Conf.logger.Info(fmt.Sprintf("%s file not load.because not exist!", *configFile))
		return
	}
	_, err := toml.Decode(*configFile, Conf)
	if err != nil {
		Conf.logger.Info(fmt.Sprintf("%s decode fail. please check format !", *configFile))
		return
	}
}
