package main

import (
	"errors"
	"log"
	"net"
	"ntn/common"
	"strings"
	"sync"
)

var reqHeaders sync.Map
var reqClients sync.Map
var httpServes sync.Map   //HTTP协议服务
var httpsServes sync.Map  //HHTPS协议服务
var clientServes sync.Map //客户端服务

func handleForwardConn(conn net.Conn, data common.JSON) {
	key := data["key"]
	headers, _ := reqHeaders.Load(key)
	reqConn, _ := reqClients.Load(key)

	conn.Write([]byte(headers.(string))) //给客户端发送请求头
	reqHeaders.Delete(key)               //完成删除客户端请求头

	common.Forward(conn, reqConn.(net.Conn)) //转发数据
	reqClients.Delete(key)                   //完成删除客户端连接
}

func handleClientConn(conn net.Conn) {
	clientMsg := new(common.Message)
	err := clientMsg.Read(conn)

	if err != nil {
		defer conn.Close()
		errMsg := common.NewMessage(common.ERROR, common.JSON{
			"msg": err.Error(),
		})

		errMsg.Send(conn)
		return
	}

	switch clientMsg.Type {
	case common.LOGIN:
		login(conn, clientMsg)
	case common.REQCONN:
		handleForwardConn(conn, clientMsg.Data)
	}
}

//获取域名
func getDomain(token string) (domain string, err error) {
	if token == "1" {
		return "", errors.New("Token验证失败")
	}

	return "localhost", nil
}

//注册用户服务逻辑
func regServe(conn net.Conn, serves []common.Serve) {
	for _, v := range serves {
		switch v.Type {
		case "http":
			println("Reg Serve Http: http://" + v.Domain + "->" + v.Addr)
			httpServes.Store(v.Domain, conn)
			clientServes.Store(v.Domain, v.Addr)
		case "https":
			println("Reg Serve Https: https://" + v.Domain + "->" + v.Addr)
			//httpsClients.Store(v.Domain, conn)
		case "tunnel":
			fallthrough
		default:
			println("Reg Serve Tunnel: tcp://" + v.Domain + "->" + v.Addr)
		}
	}
}

func isOldClient(ver string) bool {
	nowVer := strings.Replace(common.Version, ".", "", -1)
	clientVer := strings.Replace(ver, ".", "", -1)

	return clientVer < nowVer
}

//客户端登录验证逻辑
func login(conn net.Conn, msg *common.Message) {
	if isOldClient(msg.Data["version"]) {
		defer conn.Close()
		errMsg := common.NewMessage(common.FATAL, common.JSON{
			"msg": "客户端版本过低，请先升级",
		})

		errMsg.Send(conn)
		return
	}

	token := msg.Data["token"]
	_, err := getDomain(token)

	var loginMsg *common.Message
	if err != nil {
		loginMsg = common.NewMessage(common.FATAL, common.JSON{
			"msg": err.Error(),
		})

		defer conn.Close()
	} else {
		loginMsg = common.NewMessage(common.MESSAGE, common.JSON{
			"msg": "登录成功",
		})
	}

	loginMsg.Send(conn)
	regServe(conn, msg.Serves) //注册用户服务
}

func Serve(id, addr string, f func(c net.Conn)) {
	ln, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatal("服务启动失败", err)
		return
	}

	log.Println(id + "服务运行中")
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		go f(conn)
	}
}

func main() {
	go Serve("HTTP", ":80", handleHTTPConn)

	Serve("NTN", ":9188", handleClientConn)
}
