package main

import (
	"fmt"
	"net"
)

func handleTcpConn(conn net.Conn, port uint) {

}

func runTcpServe(ln net.Listener, port uint) {
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			break
		}

		go handleTcpConn(conn, port)
	}
}

func regTcpServe(client *Client, port uint) (err error) {
	ln, err := net.Listen("tcp4", fmt.Sprintf(":%v", port))

	if err == nil {
		client.TCP.Store(port, ln)
		go runTcpServe(ln, port)
	}
	return
}
