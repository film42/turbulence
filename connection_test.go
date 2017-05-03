package main

import (
	"fmt"
	"strings"
	"testing"
)

func resetCredentials() {
	AuthenticationRequired = false
	setCredentials("", "")
}

func setCredentials(user, pass string) {
	AuthenticationRequired = true
	Username = user
	Password = pass
}

func basicHttpProxyRequest() string {
	return "GET http://httpbin.org/headers HTTP/1.1\r\nHost: httpbin.org\r\nUser-Agent: curl/7.53.1\r\n\r\n"
}

func TestInvalidCredentials(t *testing.T) {
	InitLogger()
	setCredentials("test", "hello")
	incoming := NewMockConn()
	conn := NewConnection(incoming)

	go func() {
		fmt.Println("start reading")
		conn.Handle()
		fmt.Println("done reading")
	}()

	fmt.Println("start writing")
	incoming.ClientWriter.Write([]byte(basicHttpProxyRequest()))
	fmt.Println("done writing")

	fmt.Println("start clinet reading")
	buffer := make([]byte, 100)
	incoming.ClientReader.Read(buffer)
	fmt.Println("finish client reading")
	response := strings.TrimRight(string(buffer), "\x000")

	expected := "HTTP/1.0 407 Proxy authentication required\r\n\r\n"
	if response != expected {
		fmt.Println("Expected", expected, len(expected), "but found", response, len(response))
		t.Fatal()
	}
}
