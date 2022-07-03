package serves

import (
	"fmt"
	"net"
	"ntn/common"
)

func handleTcpConn(conn net.Conn, client *Client, port uint) {
	addr := conn.RemoteAddr().String()
	serve := client.GetTCPServe(port)

	if serve != nil {
		reqMsg := common.NewMessage(common.HASREQ, common.JSON{
			"key":  addr,
			"type": serve.Type,
			"addr": serve.Addr,
		})

		reqMsg.Send(client.Conn)
		reqClients.Store(addr, conn)
	} else {
		conn.Close() //没有
	}
}

func tcpServe(ln net.Listener, client *Client, port uint) {
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			break
		}

		go handleTcpConn(conn, client, port)
	}
}

func regTcpServe(client *Client, port uint) (err error) {
	ln, err := net.Listen("tcp4", fmt.Sprintf(":%v", port))

	if err == nil {
		client.TCP.Store(port, ln)
		go tcpServe(ln, client, port)
	}
	return
}
