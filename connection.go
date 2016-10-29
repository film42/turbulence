package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net"
	"net/http"
)

type connection struct {
	incoming net.Conn
	outgoing net.Conn
	proxy    proxy
	id       string
}

func (c *connection) Handle() {
	logger.Info.Println(c.id, "Handling new connection.")

	reader := bufio.NewReader(c.incoming)
	request, err := http.ReadRequest(reader)
	if err == io.EOF {
		logger.Warn.Println(c.id, "Incoming connection disconnected.")
		return
	}

	if err != nil {
		logger.Warn.Println(c.id, "Could not parse or read request from incoming connection:", err)
		return
	}

	logger.Info.Println(c.id, "Processing connection to:", request.Method, request.Host)

	if request.Method == "CONNECT" {
		c.proxy = &httpsProxy{}
	} else {
		c.proxy = &httpProxy{}
	}

	err = c.proxy.SetupOutgoing(c, request)
	if err != nil {
		logger.Warn.Println(c.id, err)
		return
	}

	// Spawn incoming->outgoing and outgoing->incoming streams.
	signal := make(chan error)
	go streamBytes(c.incoming, c.outgoing, signal)
	go streamBytes(c.outgoing, c.incoming, signal)

	// Wait for either stream to complete and finish.
	err = <-signal
	if err != nil {
		logger.Warn.Println(c.id, "Error reading or writing data", request.Host, err)
		return
	}
}

func (c *connection) Close() {
	if c.incoming != nil {
		c.incoming.Close()
	}

	if c.outgoing != nil {
		c.outgoing.Close()
	}

	logger.Info.Println(c.id, "Connection closed.")
}

func newConnectionId() string {
	bytes := make([]byte, 3) // 6 characters long.
	if _, err := rand.Read(bytes); err != nil {
		return "[ERROR-MAKING-UUID]"
	}
	return "[" + hex.EncodeToString(bytes) + "]"
}

func NewConnection(incoming net.Conn) *connection {
	return &connection{
		id:       newConnectionId(),
		incoming: incoming,
	}
}
