package utils

import (
	"log"
	"net"
	"ntn/common"
)

func SendRegServeRes(conn net.Conn, serves []common.Serve, status string) {
	resMsg := common.Message{
		Type: common.REGRES,
		Data: common.JSON{
			"status": status,
		},
		Serves: serves,
	}

	if status == common.OK {
		for _, v := range serves {
			if v.Type == common.HTTP || v.Type == common.HTTPS {
				log.Printf("Reg Serve %v: %s://%s->%s\n", status, v.Type, v.Domain, v.Addr)
			} else {
				log.Printf("Reg Serve %v: %s://%v->%s\n", status, v.Type, v.Port, v.Addr)
			}
		}
	}

	resMsg.Send(conn)
}

func SendStatusMsg(conn net.Conn, t uint, msg, status string) {
	statusMsg := common.NewMessage(t, common.JSON{
		"msg":    msg,
		"status": status,
	})

	statusMsg.Send(conn)
}

func SendMessage(conn net.Conn, t uint, msg string) {
	statusMsg := common.NewMessage(t, common.JSON{
		"msg": msg,
	})

	statusMsg.Send(conn)
}
