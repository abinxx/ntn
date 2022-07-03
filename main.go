package main

import (
	"ntn/common"
	"ntn/serves"
)

type Config struct {
	Port uint   //监听端口
	Ca   string //证书文件路径
	Key  string //私钥文件路径
}

var appConfig Config

func main() {
	common.GetConfig(&appConfig)
	serves.LoadTLsConfig(appConfig.Ca, appConfig.Key)
	serves.Start(appConfig.Port)
}
