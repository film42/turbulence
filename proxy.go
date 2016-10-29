package main

import (
	"io"
	"net"
	"net/http"
)

type proxy interface {
	Handle(*connection, *http.Request)
}

func streamBytes(src net.Conn, dest net.Conn, signal chan error) {
	_, err := io.Copy(dest, src)
	signal <- err
}
