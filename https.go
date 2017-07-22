package main

import (
	"errors"
	"fmt"
	"net/http"
)

const ConnectionEstablished = "HTTP/1.0 200 Connection established\r\n\r\n"

type httpsProxy struct{}

func (hp *httpsProxy) SetupOutgoing(connection *connection, request *http.Request) error {
	// Create our outgoing connection.
	outgoingHost := request.URL.Host
	outgoingConn, err := connection.Dial("tcp", outgoingHost)
	if err != nil {
		return errors.New(fmt.Sprint("Error opening outgoing connection to", outgoingHost, err))
	}
	connection.outgoing = outgoingConn

	// Signal to the incoming connection that the outgoing connection was established.
	_, err = connection.incoming.Write([]byte(ConnectionEstablished))
	if err != nil {
		return errors.New(fmt.Sprint("Could not send Continue response to incoming", outgoingHost, err))
	}

	return nil
}
