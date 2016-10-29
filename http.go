package main

import (
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

func (hp *httpProxy) Handle(connection *connection, request *http.Request) {
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
		fmt.Println("Error making outgoing request to", request.Host, err)
		return
	}

	connection.outgoing = outgoingConn
	err = request.Write(connection.outgoing)
	if err != nil {
		fmt.Println("Error writing request from incoming->outgoing", request.Host, err)
		return
	}

	// Spawn incoming->outgoing and outgoing->incoming streams.
	signal := make(chan error)
	go streamBytes(connection.incoming, connection.outgoing, signal)
	go streamBytes(connection.outgoing, connection.incoming, signal)

	// Wait for either stream to complete and finish.
	err = <-signal
	if err != nil {
		fmt.Println("Error reading or writing data", request.Host, err)
		return
	}
}
