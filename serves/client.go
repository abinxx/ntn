package serves

import (
	"log"
	"net"
	"ntn/common"
	"sync"
)

var clients = make(map[string]*Client)

type Client struct {
	Conn   net.Conn
	Serves []common.Serve
	TCP    sync.Map //TCP隧道服务
	UDP    sync.Map //UDP隧道服务
}

func NewClient(conn net.Conn) *Client {
	return &Client{Conn: conn}
}

func (c *Client) GetServeByDomain(t, domain string) *common.Serve {
	for _, v := range c.Serves {
		if v.Type == t && v.Domain == domain {
			return &v
		}
	}

	return nil
}

func (c *Client) GetServeByPort(t string, port uint) *common.Serve {
	for _, v := range c.Serves {
		if v.Type == "tcp" && v.Port == port {
			return &v
		}
	}

	return nil
}

func (c *Client) GetServe(domain string, isHttps bool) *common.Serve {
	if isHttps {
		return c.GetServeByDomain(common.HTTPS, domain)
	}

	return c.GetServeByDomain(common.HTTP, domain)
}

func (c *Client) GetTCPServe(port uint) *common.Serve {
	return c.GetServeByPort("tcp", port)
}

func (c *Client) GetUDPServe(port uint) *common.Serve {
	return c.GetServeByPort("udp", port)
}

func CloseAllListen(key, value interface{}) bool {
	ln, _ := value.(net.Listener)
	ln.Close()
	log.Println("Close Serve by Port:", key)
	return false
}

func (c *Client) Close() {
	c.Conn.Close()
	c.TCP.Range(CloseAllListen)
	c.UDP.Range(CloseAllListen)
}

func GetClientByConn(conn net.Conn) *Client {
	for _, v := range clients {
		if conn == v.Conn {
			return v
		}
	}
	return nil
}

func CloseClient(conn net.Conn) {
	for k, v := range clients {
		if v.Conn == conn {
			v.Close()
			delete(clients, k)
			log.Println("Now Online Clients:", len(clients))
			return
		}
	}
}
