package common

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"log"
	"net"
)

const ByteHeader = 0x66 //消息头部验证标识

type JSON map[string]string

type Message struct {
	Type   uint    `json:"type"`
	Data   JSON    `json:"data,omitempty"`
	Serves []Serve `json:"serves,omitempty"`
}

func (msg *Message) Send(conn net.Conn) (err error) {
	msgBytes, err := json.Marshal(&msg)
	log.Println("SEND Msg:", string(msgBytes))

	dataLen := uint16(len(msgBytes))
	log.Printf("SEND Msg Len: %d Byte\n", dataLen)

	buf := make([]byte, 3)
	buf[0] = ByteHeader
	binary.BigEndian.PutUint16(buf[1:3], dataLen) //将长度转大端字节

	if err == nil {
		conn.Write(buf) //发送消息长度
		conn.Write(msgBytes)
	}
	return
}

func (msg *Message) Read(conn net.Conn) (err error) {
	buf := make([]byte, 3)
	n, err := conn.Read(buf) //读取消息头部

	if err != nil {
		return
	} else if n < 3 || buf[0] != ByteHeader { //未知的客户端
		return errors.New("Error Client Addr: " + conn.RemoteAddr().String())
	}

	dataLen := uint16(binary.BigEndian.Uint16(buf[1:n])) //解析消息长度
	log.Printf("READ Msg Len: %d Byte\n", dataLen)

	msgBtyes := make([]byte, dataLen)
	n, err = conn.Read(msgBtyes) //读取消息

	log.Println("READ Msg:", string(msgBtyes[:n]))
	err = json.Unmarshal(msgBtyes[:n], &msg)
	return
}

func NewMessage(t uint, data JSON) *Message {
	return &Message{
		Type: t,
		Data: data,
	}
}
