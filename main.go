package main

import (
	"bufio"
	"errors"
	"io"
	"net"
	"ntn/common"
	"strings"
	"sync"
)

var reqHeaders sync.Map
var reqClients sync.Map
var clients sync.Map

func getHTTPHeaders(conn net.Conn) (host string) {
	addr := conn.RemoteAddr().String()
	var headers []string
	buf := bufio.NewReader(conn)

	for {
		line, err := buf.ReadString('\n')

		if host == "" {
			if strings.Contains(line, "Host") || strings.Contains(line, "host") {
				hostArr := strings.Split(line, ":")
				if len(hostArr) > 1 { //格式Host: 127.0.0.1:80
					host = strings.TrimSpace(hostArr[1])
				}
			}
		}

		headers = append(headers, line) //追加请求头

		if line == "\r\n" || err == io.EOF {
			data, _ := buf.Peek(buf.Buffered())
			headers = append(headers, string(data)) //追究POST数据
			break
		}
	}

	reqHeaders.Store(addr, headers)
	return
}

func handleHTTPConn(conn net.Conn) {
	addr := conn.RemoteAddr().String()
	println("New HTTP CONN:", addr)

	host := getHTTPHeaders(conn)
	if clientConn, ok := clients.Load(host); ok {
		reqMsg := common.NewMessage(common.HASREQ, common.JSON{
			"key": addr,
		})

		reqMsg.Send(clientConn.(net.Conn)) //通知客户端有新请求
		reqClients.Store(addr, conn)       //保存HTTP请求连接
	} else {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\nClient is't Online."))
		conn.Close()
	}
}

func handleForwardConn(conn net.Conn, data common.JSON) {
	key := data["key"]
	headers, _ := reqHeaders.Load(key)
	reqConn, _ := reqClients.Load(key)

	header := strings.Join(headers.([]string), "")
	conn.Write([]byte(header)) //给客户端发送请求头
	reqHeaders.Delete(key)

	common.Copy(conn, reqConn.(net.Conn)) //转发数据
}

func handleClientConn(conn net.Conn) {
	connectMsg := new(common.Message)
	err := connectMsg.Read(conn)

	if err != nil {
		defer conn.Close()
		errMsg := common.NewMessage(common.ERROR, common.JSON{
			"msg": err.Error(),
		})

		errMsg.Send(conn)
		return
	}

	switch connectMsg.Type {
	case common.LOGIN:
		login(conn, connectMsg.Data)
	case common.REQCONN:
		handleForwardConn(conn, connectMsg.Data)
	}
}

//获取域名
func getDomain(data common.JSON) (domain string, err error) {
	token := data["token"]

	if token == "1" {
		return "", errors.New("Token验证失败")
	}

	if domain, ok := data["domain"]; ok {
		return domain, nil
	}

	return "nsp.bincs.cn", nil
}

//客户端登录验证逻辑
func login(conn net.Conn, data common.JSON) {
	domain, err := getDomain(data)

	var loginMsg *common.Message
	if err != nil {
		loginMsg = common.NewMessage(common.FATAL, common.JSON{
			"msg": err.Error(),
		})

		defer conn.Close()
	} else {
		clients.Store(domain, conn)
		loginMsg = common.NewMessage(common.MESSAGE, common.JSON{
			"msg":    "登录成功",
			"domain": domain,
		})
	}

	loginMsg.Send(conn)
}

func Serve(id, addr string, f func(c net.Conn)) {
	ln, err := net.Listen("tcp", addr)

	if err != nil {
		println("服务启动失败", err)
		return
	}

	println(id + "服务运行中")
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
