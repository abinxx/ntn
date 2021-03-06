package serves

import (
	"log"
	"net"
	"ntn/common"
	"ntn/utils"
	"strings"
)

func handleForwardConn(conn net.Conn, data common.JSON) {
	key := data["key"]
	defer conn.Close()
	defer reqClients.Delete(key) //完成删除连接

	if data["type"] == common.HTTP || data["type"] == common.HTTPS {
		headers, _ := reqHeaders.Load(key)
		conn.Write([]byte(headers.(string))) //发送请求头
		reqHeaders.Delete(key)               //完成删除保存的请求头
		up := (int64)(len(headers.(string)))
		common.OnByte(up, 0) //请求头数据上传长度
	}

	if reqConn, ok := reqClients.Load(key); ok {
		common.Forward(conn, reqConn.(net.Conn)) //转发数据
	}
}

func handleClientConn(conn net.Conn) {
	defer conn.Close()
	log.Printf("New Client Conn: %s\n", conn.RemoteAddr().String())
	clientMsg := new(common.Message)

	for {
		err := clientMsg.Read(conn)
		if err != nil {
			log.Println("Client Disconnect:", err.Error())
			CloseClient(conn)
			return
		}

		switch clientMsg.Type {
		case common.LOGIN:
			login(conn, clientMsg)
		case common.REGSERVE:
			regServe(conn, clientMsg.Serves)
		case common.TUNNEL:
			handleForwardConn(conn, clientMsg.Data)
			return //隧道结束就退出
		}
	}
}

//注册服务逻辑
func regServe(conn net.Conn, serves []common.Serve) {
	client := GetClientByConn(conn) //获取保存的客户端

	if client == nil {
		conn.Close() //客户端不存在 关闭注册连接
		return
	}

	var err error
	var errServes []common.Serve
OuterLoop:
	for _, v := range serves {
		switch v.Type {
		case common.HTTP:
			fallthrough
		case common.HTTPS:
			err = regHttpAndHttpsServe(v.Domain, v.Type == common.HTTPS)
		case common.TCP:
			err = regTcpServe(client, v.Port)
		case common.UDP:
			//log.Printf("Reg Serve UDP: %v->%v", v.Port, v.Addr)
			break OuterLoop
		default:
			log.Println("Serve Type Error")
			break OuterLoop
		}

		if err != nil {
			log.Printf("Reg %s Serve: %s", v.Type, err.Error())
			errServes = append(errServes, v)
		} else {
			client.Serves = append(client.Serves, v)
		}
	}

	utils.SendRegServeRes(conn, client.Serves, common.OK)
	utils.SendRegServeRes(conn, errServes, common.NO)
}

func isOldClient(ver string) bool {
	nowVer := strings.Replace(common.Version, ".", "", -1)
	clientVer := strings.Replace(ver, ".", "", -1)

	return clientVer < nowVer
}

//客户端登录验证逻辑
func login(conn net.Conn, msg *common.Message) {
	if isOldClient(msg.Data["version"]) {
		defer conn.Close()
		errMsg := common.NewMessage(common.FATAL, common.JSON{
			"msg": "客户端版本过低，请先升级",
		})

		errMsg.Send(conn)
		return
	}

	token := msg.Data["token"]
	//_, err := getDomain(token) //插件登录验证 返回成功失败
	var err error

	if err != nil {
		utils.SendStatusMsg(conn, common.LOGIN, err.Error(), common.NO)
	} else {
		utils.SendStatusMsg(conn, common.LOGIN, "身份验证成功", common.OK)
	}

	if client, ok := clients[token]; ok {
		utils.SendMessage(client.Conn, common.FATAL, "账号在新设备上登录")
		client.Close()
		delete(clients, token)
	}

	clients[token] = NewClient(conn) //保存客户端
	log.Println("Now Online Clients:", len(clients))
}
