package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
)

const ConnectionEstablished = "HTTP/1.0 200 Connection established\r\n\r\n"

type httpsProxy struct{}

func (hp *httpsProxy) Handle(connection *connection, request *http.Request) {
	// Create our outgoing connection.
	outgoingHost := request.URL.Host
	outgoingConn, err := net.Dial("tcp", outgoingHost)
	if err != nil {
		fmt.Println("Error opening outgoing connection to", outgoingHost, err)
		return
	}
	connection.outgoing = outgoingConn

	// Signal to the incoming connection that the outgoing connection was established.
	_, err = connection.incoming.Write([]byte(ConnectionEstablished))
	if err != nil {
		fmt.Println("Could not send Continue response to client: " + err.Error())
	}

	// Spawn incoming->outgoing and outgoing->incoming streams.
	signal := make(chan bool)
	go streamBytes(connection.incoming, connection.outgoing, signal)
	go streamBytes(connection.outgoing, connection.incoming, signal)

	// Wait for either stream to complete and finish.
	<-signal
}

func streamBytes(src net.Conn, dest net.Conn, signal chan bool) {
	// Read until EOF or error.
	io.Copy(dest, src)
	signal <- true
}
