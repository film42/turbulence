package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

type connection struct {
	incoming net.Conn
	outgoing net.Conn
	proxy    proxy
}

func (c *connection) Handle() {
	reader := bufio.NewReader(c.incoming)
	request, err := http.ReadRequest(reader)
	if err != nil {
		fmt.Println("Could not parse or read request from incoming connection:", err)
		return
	}

	if request.Method == "CONNECT" {
		c.proxy = &httpsProxy{}
	} else {
		c.proxy = &httpProxy{}
	}

	c.proxy.Handle(c, request)
}

func (c *connection) Close() {
	if c.incoming != nil {
		c.incoming.Close()
	}

	if c.outgoing != nil {
		c.outgoing.Close()
	}
}

func NewConnection(incoming net.Conn) *connection {
	return &connection{
		incoming: incoming,
	}
}
