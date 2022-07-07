package serves

import (
	"fmt"
	"log"
	"net"
	"sync"
)

var reqHeaders sync.Map
var reqClients sync.Map

func Serve(id, addr string, f func(net.Conn)) {
	ln, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatal("Run Serve Error:", err)
	}

	log.Println(id + " Serve Runing...")
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		go f(conn)
	}
}

func Start(port uint) {
	go Serve("HTTP", ":80", handleHTTPConn)
	go Serve("HTTPS", ":443", handleHTTPSConn)

	Serve("NTN", fmt.Sprintf(":%v", port), handleClientConn)
}
