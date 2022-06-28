package common

import (
	"fmt"
	"net"
)

const BR = "\r\n"

func SendHttpRes(conn net.Conn, msg string) {
	defer conn.Close()
	fmt.Fprintf(conn, "HTTP/1.1 200 OK%sContent-length: %v%v", BR, len(msg), BR+BR+msg)
}
