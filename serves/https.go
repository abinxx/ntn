package serves

import (
	"crypto/tls"
	"errors"
	"log"
	"net"
)

var config *tls.Config

func LoadTLsConfig(ca, key string) {
	cer, err := tls.LoadX509KeyPair(ca, key)
	if err != nil {
		log.Fatalln("Load CA Error:", err)
	}
	config = &tls.Config{Certificates: []tls.Certificate{cer}}
}

func handleHTTPSConn(conn net.Conn) {
	tlsConn := tls.Server(conn, config) //转成TLS连接
	handleHttpAndHttps(tlsConn, true)
}

func regHttpAndHttpsServe(domain string, isHttps bool) (err error) {
	for _, v := range clients {
		serve := v.GetServe(domain, isHttps)
		if serve != nil {
			return errors.New(domain + "已被使用")
		}
	}
	return
}
