package common

import (
	"flag"
	"fmt"

	"github.com/spf13/viper"
)

type Serve struct {
	Addr   string `json:"addr"`             //内网地址
	Type   string `json:"type"`             //类型
	Domain string `json:"domain,omitempty"` //域名
}

func GetConfig(appConfig interface{}) {
	var config string

	flag.StringVar(&config, "c", "", "Config file path.")
	flag.Parse()

	if config == "" {
		config = "config.yaml" //默认配置文件路径
	}
	v := viper.New()
	v.SetConfigFile(config)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Read Config: %s\n", err))
	}

	if err := v.Unmarshal(&appConfig); err != nil {
		panic(fmt.Errorf("Config ERROR: %s\n", err))
	}
}
