package main

import (
	"bufio"
	"fmt"
	"io"
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
	if err == io.EOF {
		fmt.Println("Incoming connection disconnected.")
		return
	}

	if err != nil {
		fmt.Println("Could not parse or read request from incoming connection:", err)
		return
	}

	if request.Method == "CONNECT" {
		c.proxy = &httpsProxy{}
	} else {
		c.proxy = &httpProxy{}
	}

	err = c.proxy.SetupOutgoing(c, request)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Spawn incoming->outgoing and outgoing->incoming streams.
	signal := make(chan error)
	go streamBytes(c.incoming, c.outgoing, signal)
	go streamBytes(c.outgoing, c.incoming, signal)

	// Wait for either stream to complete and finish.
	err = <-signal
	if err != nil {
		fmt.Println("Error reading or writing data", request.Host, err)
	}
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
