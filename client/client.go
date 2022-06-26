package main

import (
	"fmt"
	"net"
	"ntn/common"
	"os"
)

const (
	remoteAddr = "120.79.182.242:9188"
	localAddr  = "192.168.2.100:80"
)

func handleForwardConn(data common.JSON) {
	remoteConn, _ := net.Dial("tcp", remoteAddr)

	httpMsg := common.NewMessage(common.REQCONN, common.JSON{
		"key": data["key"],
	})

	httpMsg.Send(remoteConn)

	localConn, _ := net.Dial("tcp", localAddr)

	common.Copy(remoteConn, localConn)
}

func login(conn net.Conn) {

	connectMsg := common.NewMessage(common.LOGIN, common.JSON{
		"token":  "aaaaaaaaaaaaaaa",
		"domain": "home.nsp.bincs.cn",
	})

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
			fmt.Println(msg)
			os.Exit(1)
		}
	}
}

func main() {
	conn, err := net.Dial("tcp", remoteAddr)

	if err != nil {
		println("连接服务器失败", err.Error())
		return
	}

	login(conn) //登录验证Token
	serve(conn) //运行主服务
}
