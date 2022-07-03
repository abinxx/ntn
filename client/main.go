package main

import (
	"crypto/tls"
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
		"key":  data["key"],
		"type": data["type"],
	})

	httpMsg.Send(remoteConn)

	var localConn net.Conn
	if data["type"] == common.HTTPS {
		localConn, err = tls.Dial("tcp", data["addr"], &tls.Config{InsecureSkipVerify: true})
	} else {
		localConn, err = net.Dial("tcp", data["addr"])
	}

	if err != nil {
		log.Println("连接本地服务失败:", err.Error())
		common.SendHttpRes(remoteConn, "Local serve connection error.")
		return
	}

	common.Forward(remoteConn, localConn)
}

//登录验证
func connect(conn net.Conn) {
	connectMsg := common.NewMessage(common.LOGIN, common.JSON{
		"token":   appConfig.Token,
		"version": common.Version,
		"os":      runtime.GOOS,
	})

	connectMsg.Send(conn)
}

//登录结果
func loginRes(conn net.Conn, data common.JSON) {
	if data["status"] == common.OK {
		regRemoteServe(conn, appConfig.Serves)
	} else {
		log.Fatal(data["msg"])
	}
}

//向服务器注册服务
func regRemoteServe(conn net.Conn, serves []common.Serve) {
	regMsg := common.Message{
		Type:   common.REGSERVE,
		Serves: serves,
	}

	regMsg.Send(conn)
}

//注册服务结果
func regRemoteServeRes(msg *common.Message) {
	data := msg.Data

	for _, v := range msg.Serves {
		if v.Type == common.HTTP || v.Type == common.HTTPS {
			log.Printf("Reg Serve %v: %s://%s->%s\n", data["status"], v.Type, v.Domain, v.Addr)
		} else {
			log.Printf("Reg Serve %v: %s://%v->%s\n", data["status"], v.Type, v.Port, v.Addr)
		}
	}
}

func serve(conn net.Conn) {
	defer conn.Close()
	conn.(*net.TCPConn).SetKeepAlive(true) //开启保持长连接

	for {
		msg := new(common.Message)
		err := msg.Read(conn)
		if err != nil {
			log.Println("服务器已断开连接:", err.Error())
			break
		}

		switch msg.Type {
		case common.LOGINRES:
			loginRes(conn, msg.Data)
		case common.REGRES:
			go regRemoteServeRes(msg)
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
