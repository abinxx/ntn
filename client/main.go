package main

import (
	"log"
	"ntn/common"
	"time"
)

type Config struct {
	Server string
	Token  string
	Sleep  uint
	Serves []common.Serve
}

var appConfig Config

func main() {
	common.GetConfig(&appConfig)

	for {
		Start(appConfig.Server)
		time.Sleep(time.Second * time.Duration(appConfig.Sleep))
		log.Println("正在尝试重连服务器...")
	}
}
