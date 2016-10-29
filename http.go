package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

type httpProxy struct{}

// Right now URL Host is a host or host:port. This snippet is taken from a future
// version of go.
// https://github.com/golang/go/commit/1ff19201fd898c3e1a0ed5d3458c81c1f062570b
func portOnly(hostport string) string {
	colon := strings.IndexByte(hostport, ':')
	if colon == -1 {
		return ""
	}
	if i := strings.Index(hostport, "]:"); i != -1 {
		return hostport[i+len("]:"):]
	}
	if strings.Contains(hostport, "]") {
		return ""
	}
	return hostport[colon+len(":"):]
}

func (hp *httpProxy) SetupOutgoing(connection *connection, request *http.Request) error {
	// Connect to outgoing host and write request bytes.
	request.RequestURI = ""

	// Port may not be set, but default is always 80
	host := request.URL.Host
	port := portOnly(host)
	if port == "" {
		host += ":80"
	}

	// Create our outgoing connection.
	outgoingConn, err := net.Dial("tcp", host)
	if err != nil {
		return errors.New(fmt.Sprint("Error making outgoing request to", request.Host, err))
	}

	connection.outgoing = outgoingConn
	err = request.Write(connection.outgoing)
	if err != nil {
		return errors.New(fmt.Sprint("Error writing request from incoming->outgoing", request.Host, err))
	}

	return nil
}
