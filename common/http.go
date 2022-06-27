package common

import "net"

func SendHttpRes(conn net.Conn, msg string) {
	defer conn.Close()
	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n" + msg))
}
