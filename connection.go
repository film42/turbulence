package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
)

type connection struct {
	id       string
	incoming net.Conn
	outgoing net.Conn
	proxy    proxy
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

	defer request.Body.Close()

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
	signal := make(chan error, 1)
	go streamBytes(c.incoming, c.outgoing, signal)
	go streamBytes(c.outgoing, c.incoming, signal)

	// Wait for either stream to complete and finish. The second will always be an error.
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
		return "[ERROR-MAKING-ID]"
	}
	return "[" + hex.EncodeToString(bytes) + "]"
}

func NewConnection(incoming net.Conn) *connection {
	newId := fmt.Sprint(newConnectionId(), " [", incoming.RemoteAddr().String(), "]")

	return &connection{
		id:       newId,
		incoming: incoming,
	}
}
