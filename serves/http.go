package serves

import (
	"bufio"
	"io"
	"log"
	"net"
	"ntn/common"
	"strings"
)

//解析HTTP头部Host
func GetHostWithHeaders(conn net.Conn) (host string) {
	addr := conn.RemoteAddr().String()
	var headers string
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

		headers += line //追加请求头

		if line == "\r\n" || err == io.EOF {
			data, _ := buf.Peek(buf.Buffered())
			headers += string(data) //追加POST数据
			break
		}
	}

	reqHeaders.Store(addr, headers)
	return
}

func handleHttpAndHttps(conn net.Conn, isHttps bool) {
	addr := conn.RemoteAddr().String()
	log.Println("New Conn:", addr)

	host := GetHostWithHeaders(conn)

	for _, v := range clients {
		serve := v.GetServe(host, isHttps)

		if serve != nil {
			reqMsg := common.NewMessage(common.HASREQ, common.JSON{
				"key":  addr,
				"type": serve.Type,
				"addr": serve.Addr,
			})

			reqMsg.Send(v.Conn)          //通知客户端有新请求
			reqClients.Store(addr, conn) //保存HTTP请求连接
			return
		}
	}

	common.SendHttpRes(conn, "Client is't online.")
}

func handleHTTPConn(conn net.Conn) {
	handleHttpAndHttps(conn, false)
}