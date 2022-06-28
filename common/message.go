package common

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"net"
)

type JSON map[string]string

type Message struct {
	Type   uint    `json:"type"`
	Data   JSON    `json:"data"`
	Serves []Serve `json:"serves,omitempty"`
}

func (msg *Message) Send(conn net.Conn) (err error) {
	msgBytes, err := json.Marshal(&msg)
	log.Println("Send Msg:", string(msgBytes))

	dataLen := uint16(len(msgBytes))
	log.Printf("Send Msg Len: %d Byte\n", dataLen)

	lenBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBytes, dataLen) //将长度转大端字节

	if err == nil {
		conn.Write(lenBytes) //发送消息长度
		conn.Write(msgBytes)
	}
	return
}

func (msg *Message) Read(conn net.Conn) (err error) {
	buf := make([]byte, 2)
	_, err = conn.Read(buf)

	if err != nil {
		return //读取失败
	}

	dataLen := uint16(binary.BigEndian.Uint16(buf)) //解析消息长度
	log.Printf("Read Msg Len: %d Byte\n", dataLen)

	msgBtyes := make([]byte, dataLen)
	_, err = conn.Read(msgBtyes) //读取消息

	log.Println("Read Msg:", string(msgBtyes))
	err = json.Unmarshal(msgBtyes, &msg)
	return
}

func NewMessage(t uint, data JSON) *Message {
	return &Message{
		Type: t,
		Data: data,
	}
}
