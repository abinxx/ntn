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
	defer conn.Close()
	clientMsg := new(common.Message)

	for {
		err := clientMsg.Read(conn)
		if err != nil {
			log.Println("Client Disconnect:", err.Error())
			CloseClient(conn)
			return
		}

		switch clientMsg.Type {
		case common.LOGIN:
			login(conn, clientMsg)
		case common.TUNNEL:
			handleForwardConn(conn, clientMsg.Data)
			break //隧道结束就退出
		}
	}
}

//获取域名
func getDomain(token string) (domain string, err error) {
	if token == "1" {
		return "", errors.New("Token验证失败")
	}

	return "localhost", nil
}

//注册服务逻辑
func regServe(conn net.Conn, serves []common.Serve) *Client {
	client := NewClient(conn)

	for _, v := range serves {
		resStatus := false
		switch v.Type {
		case "http":
			resStatus = true
			log.Println("Reg Serve Http: http://" + v.Domain + "->" + v.Addr)
		case "https":
			log.Println("Reg Serve Https: https://" + v.Domain + "->" + v.Addr)
		case "tcp":
			err := regTcpServe(client, v.Port)
			if err != nil {
				log.Printf("Reg Serve TCP: %s", err.Error())
				break
			}
			resStatus = true
			log.Printf("Reg Serve TCP: %v->%v", v.Port, v.Addr)
		case "udp":
			log.Printf("Reg Serve UDP: %v->%v", v.Port, v.Addr)
		default:
			log.Println("Reg Serve Error")
		}

		if resStatus {
			client.Serves = append(client.Serves, v)
		}
	}

	return client
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
	if client, ok := clients[token]; ok {
		clientMsg := common.NewMessage(common.FATAL, common.JSON{
			"msg": "账号在新设备上登录",
		})

		clientMsg.Send(client.Conn)
		client.Close()
		delete(clients, token)
	}
	clients[token] = regServe(conn, msg.Serves) //注册用户服务
	log.Println("Now Online Clients:", len(clients))
}

func Serve(id, addr string, f func(net.Conn)) {
	ln, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatal("服务启动失败", err)
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
