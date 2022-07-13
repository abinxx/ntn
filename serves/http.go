package serves

import (
	"bufio"
	"io"
	"log"
	"net"
	"ntn/common"
	"strings"
)

//获取HTTP头部和Host
func GetHeadersAndHost(conn net.Conn) (headers, host string) {
	buf := bufio.NewReader(conn)

	for {
		line, err := buf.ReadString('\n')

		if host == "" {
			if strings.Contains(line, "Host") || strings.Contains(line, "host") {
				hostArr := strings.Split(line, ":")
				if len(hostArr) < 1 {
					return //解析Host失败 格式->Host:ntn.bincs.cn:80
				}
				host = strings.TrimSpace(hostArr[1])
			}
		}

		headers += line //追加请求头

		if line == "\r\n" || err == io.EOF {
			data, _ := buf.Peek(buf.Buffered())
			headers += string(data) //追加POST数据
			break
		} else if err != nil {
			log.Println("Read Headers Error:", err)
			break //读取失败
		}
	}

	return
}

func handleHttp(conn net.Conn, isHttps bool) {
	//log.Println("New Http Conn:", conn.RemoteAddr().String())
	addr := conn.RemoteAddr().String()
	headers, host := GetHeadersAndHost(conn)

	for _, v := range clients {
		serve := v.GetServe(host, isHttps)

		if serve != nil {
			reqMsg := common.NewMessage(common.HASREQ, common.JSON{
				"key":  addr,
				"type": serve.Type,
				"addr": serve.Addr,
			})

			reqHeaders.Store(addr, headers) //保存HTTP请求头
			reqClients.Store(addr, conn)    //保存HTTP请求连接
			reqMsg.Send(v.Conn)             //通知客户端有新请求
			return
		}
	}

	common.SendHttpRes(conn, "Client is't online.")
}

func handleHTTPConn(conn net.Conn) {
	handleHttp(conn, false)
}
