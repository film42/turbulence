package main

import (
	"io"
	"net"
	"net/http"
)

type proxy interface {
	SetupOutgoing(*connection, *http.Request) error
}

func streamBytes(src net.Conn, dest net.Conn, signal chan error) {
	buffer := make([]byte, 1024)
	_, err := io.CopyBuffer(dest, src, buffer)
	signal <- err
}
