package common

import (
	"io"
	"log"
	"net"
)

const (
	HTTP    = "http"
	HTTPS   = "https"
	TCP     = "tcp"
	UDP     = "udp"
	OK      = "OK"    //成功
	NO      = "ERROR" //失败
	Version = "0.2.2" //当前版本号
)

const (
	LOGIN    = iota + 1 //客户端登录
	REGSERVE            //注册服务
	REGRES              //注册服务结果
	MESSAGE             //消息通知
	HASREQ              //有新连接
	TUNNEL              //新连接隧道
	ERROR               //出现错误
	FATAL               //致命错误 退出程序
)

func Forward(dst, src net.Conn) {
	defer dst.Close() //拷贝完立即释放连接
	defer src.Close() //让上传协程退出
	down := make(chan int64, 1)

	go func() {
		defer src.Close() //拷贝完立即释放连接
		defer dst.Close() //让下载携程退出

		n, err := io.Copy(dst, src)
		if err != nil {
			log.Println("Upload Error:", err.Error())
		}
		down <- n //通知上传流量
	}()

	up, err := io.Copy(src, dst)
	if err != nil {
		log.Println("Dodwload Error:", err.Error())
	}
	OnByte(<-down, up)
}

func OnByte(up, down int64) {
	log.Printf("Upload: %d Byte\tDodwload: %d Byte\n", up, down)
}
