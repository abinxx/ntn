package common

import (
	"fmt"
	"io"
	"net"
)

const (
	LOGIN   = iota //客户端登录
	MESSAGE        //普通消息通知
	HASREQ         //有新连接
	REQCONN        //新连接隧道
	ERROR          //出现错误
	FATAL          //致命错误 退出程序
)

func Copy(dst net.Conn, src net.Conn) {
	go func() {
		defer dst.Close()
		defer src.Close()

		n, err := io.Copy(dst, src)
		if err != nil {
			println(err.Error())
		}
		fmt.Printf("Upload: %d Byte\n", n)
	}()

	n, err := io.Copy(src, dst)
	if err != nil {
		println(err.Error())
	}
	fmt.Printf("Dodwload: %d Byte\n", n)
}
