package main

import (
	"crypto/tls"
	"errors"
	"log"
	"net"
	"ntn/common"
	"runtime"
)

var Client net.Conn

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

	common.Forward(localConn, remoteConn)
}

//登录验证
func login(conn net.Conn) (err error) {
	connectMsg := common.NewMessage(common.LOGIN, common.JSON{
		"token":   appConfig.Token,
		"version": common.Version,
		"os":      runtime.GOOS,
	})

	connectMsg.Send(conn)

	msg := new(common.Message)
	err = msg.Read(conn) //读取登录结果消息

	if err != nil {
		return //读取消息失败
	}

	if msg.Data["status"] != common.OK {
		err = errors.New(msg.Data["msg"])
	}
	return
}

//向服务器注册服务
func regRemoteServe(serves []common.Serve) {
	regMsg := common.Message{
		Type:   common.REGSERVE,
		Serves: serves,
	}

	regMsg.Send(Client)
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

//当有消息时 根据消息类型处理
func onMessage(msgType uint, msg *common.Message) {
	switch msgType {
	case common.REGRES:
		regRemoteServeRes(msg)
	case common.MESSAGE:
		log.Println(msg.Data["msg"])
	case common.HASREQ:
		handleForwardConn(msg.Data)
	case common.ERROR:
		log.Println(msg.Data["msg"])
	case common.FATAL:
		log.Fatal(msg.Data["msg"])
	}
}

func serve(conn net.Conn) {
	defer conn.Close()
	conn.(*net.TCPConn).SetKeepAlive(true) //开启保持长连接
	regRemoteServe(appConfig.Serves)       //注册本地服务

	for {
		msg := new(common.Message)
		err := msg.Read(conn)
		if err != nil {
			log.Println("服务器已断开连接:", err.Error())
			break
		}

		go onMessage(msg.Type, msg)
	}
}

//停止服务
func Stop() {
	Client.Close()
}

func Start(addr string) {
	conn, err := net.Dial("tcp", addr)

	if err != nil {
		log.Println("连接服务器失败:", err.Error())
		return
	}

	Client = conn //保存连接
	login(conn)   //登录验证Token
	if err != nil {
		log.Fatalln("Login Error:", err)
	}
	serve(conn) //运行主服务
}
