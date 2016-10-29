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
	_, err := io.Copy(dest, src)
	signal <- err
}
