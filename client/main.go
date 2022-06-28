package main

import (
	"fmt"
	"net"
	"ntn/common"
	"os"
)

type Config struct {
	Server string
	Token  string
	Serves []common.Serve
}

var appConfig Config

func handleForwardConn(data common.JSON) {
	remoteConn, err := net.Dial("tcp", appConfig.Server)

	if err != nil {
		remoteConn.Close()
		return
	}

	httpMsg := common.NewMessage(common.REQCONN, common.JSON{
		"key": data["key"],
	})

	httpMsg.Send(remoteConn)

	localConn, err := net.Dial("tcp", data["addr"])

	if err != nil {
		println(err.Error())
		common.SendHttpRes(remoteConn, "Local serve connection error.")
		return
	}

	common.Forward(remoteConn, localConn)
}

func login(conn net.Conn) {
	connectMsg := common.Message{
		Type: common.LOGIN,
		Data: common.JSON{
			"token":   appConfig.Token,
			"version": common.Version,
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
			println("服务器已断开连接", err.Error())
			break
		}

		switch msg.Type {
		case common.MESSAGE:
			//fmt.Println(msg.Data["msg"])
		case common.HASREQ:
			go handleForwardConn(msg.Data)
		case common.FATAL:
			fmt.Println(msg.Data["msg"])
			os.Exit(1)
		}
	}
}

func main() {
	common.GetConfig(&appConfig)
	conn, err := net.Dial("tcp", appConfig.Server)

	if err != nil {
		println("连接服务器失败", err.Error())
		return
	}

	login(conn) //登录验证Token
	serve(conn) //运行主服务
}
