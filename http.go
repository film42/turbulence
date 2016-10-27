package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

type httpProxy struct{}

func (hp *httpProxy) Handle(connection *connection, request *http.Request) {
	// Connect to outgoing host and write request bytes.
	request.RequestURI = ""
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Error making outgoing request:", err)
		return
	}
	defer response.Body.Close()

	// Dump response struct back into bytes.
	responseBytes, err := httputil.DumpResponse(response, true)
	if err != nil {
		fmt.Println("Error dumping response to write outgoung->incoming:", err)
		return
	}

	// Write back to incoming connection and finish.
	connection.incoming.Write(responseBytes)
}
