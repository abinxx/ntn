package main

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

func (c *Client) GetHTTPServe(domain string) *common.Serve {
	return c.GetServeByDomain("http", domain)
}

func (c *Client) GetHTTPSServe(domain string) *common.Serve {
	return c.GetServeByDomain("https", domain)
}

func (c *Client) GetTCPServe(port uint) *common.Serve {
	return c.GetServeByPort("tcp", port)
}

func (c *Client) GetUDPServe(port uint) *common.Serve {
	return c.GetServeByPort("udp", port)
}

func (c *Client) Close() {
	c.Conn.Close()
}

func CloseClient(conn net.Conn) {
	for k, v := range clients {
		if v.Conn == conn {
			v.Close()
			delete(clients, k)
			log.Println("Now Clients:", len(clients))
			return
		}
	}
}
