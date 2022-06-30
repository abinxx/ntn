package main

import (
	"log"
	"net"
	"ntn/common"
	"runtime"
	"time"
)

type Config struct {
	Server string
	Token  string
	Sleep  uint
	Serves []common.Serve
}

var appConfig Config

func handleForwardConn(data common.JSON) {
	remoteConn, err := net.Dial("tcp", appConfig.Server)

	if err != nil {
		log.Println("连接服务器失败:", err.Error())
		return
	}

	httpMsg := common.NewMessage(common.TUNNEL, common.JSON{
		"key": data["key"],
	})

	httpMsg.Send(remoteConn)

	localConn, err := net.Dial("tcp", data["addr"])

	if err != nil {
		log.Println("连接本地服务失败:", err.Error())
		common.SendHttpRes(remoteConn, "Local serve connection error.")
		return
	}

	common.Forward(remoteConn, localConn)
}

func connect(conn net.Conn) {
	connectMsg := common.Message{
		Type: common.LOGIN,
		Data: common.JSON{
			"token":   appConfig.Token,
			"version": common.Version,
			"os":      runtime.GOOS,
		},
		Serves: appConfig.Serves,
	}

	connectMsg.Send(conn)
}

func serve(conn net.Conn) {
	defer conn.Close()                     //关闭连接
	conn.(*net.TCPConn).SetKeepAlive(true) //保持连接

	for {
		msg := new(common.Message)
		err := msg.Read(conn)
		if err != nil {
			log.Println("服务器已断开连接:", err.Error())
			break
		}

		switch msg.Type {
		case common.MESSAGE:
			//log.Println(msg.Data["msg"])
		case common.HASREQ:
			go handleForwardConn(msg.Data)
		case common.ERROR:
			log.Println(msg.Data["msg"])
		case common.FATAL:
			log.Fatal(msg.Data["msg"])
		}
	}
}

func main() {
	common.GetConfig(&appConfig)

	for {
		conn, err := net.Dial("tcp", appConfig.Server)

		if err != nil {
			log.Println("连接服务器失败:", err.Error())
		} else {
			connect(conn) //登录验证Token
			serve(conn)   //运行主服务
		}

		time.Sleep(time.Second * time.Duration(appConfig.Sleep))
		log.Println("正在尝试重连服务器...")
	}
}
