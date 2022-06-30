package common

import (
	"flag"
	"log"

	"github.com/spf13/viper"
)

type Serve struct {
	Addr   string `json:"addr"`             //内网地址
	Type   string `json:"type"`             //类型 -http -https -tcp -udp
	Domain string `json:"domain,omitempty"` //域名
	Port   uint   `json:"port,omitempty"`   //服务端口
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
		log.Fatalf("Read Config: %s\n", err)
	}

	if err := v.Unmarshal(&appConfig); err != nil {
		log.Fatalf("Config ERROR: %s\n", err)
	}
}
