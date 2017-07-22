package main

import (
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
)

func createHttpServer(address, payload string, code int) *http.Server {
	mux := http.NewServeMux()
	server := &http.Server{Addr: address, Handler: mux}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Date", "FAKE")
		w.WriteHeader(code)
		w.Write([]byte(payload))
	})

	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}

	go server.Serve(listener)

	return server
}

func resetCredentials() {
	config.Credentials = []Credential{}
}

func setCredentials(user, pass string) {
	config.Credentials = []Credential{{Username: user, Password: pass}}
}

func basicHttpProxyRequest() string {
	return "GET http://httpbin.org/headers HTTP/1.1\r\nHost: httpbin.org\r\n\r\n"
}

func readMessage(reader io.Reader) string {
	buffer := make([]byte, 1024)
	reader.Read(buffer)
	response := strings.TrimRight(string(buffer), "\x000")
	return response
}

func TestMain(m *testing.M) {
	config = &Config{}
	InitNullLogger()
	m.Run()
}

func TestInvalidCredentials(t *testing.T) {
	setCredentials("test", "hello")
	defer resetCredentials()

	incoming := NewMockConn()
	defer incoming.CloseClient()
	conn := NewConnection(incoming)
	go conn.Handle()

	incoming.ClientWriter.Write([]byte(basicHttpProxyRequest()))

	buffer := make([]byte, 100)
	incoming.ClientReader.Read(buffer)
	response := strings.TrimRight(string(buffer), "\x000")

	expected := "HTTP/1.0 407 Proxy authentication required\r\n\r\n"
	if response != expected {
		t.Fatalf("Expected '%s' but got '%s'", expected, response)
	}
}

func TestSampleProxy(t *testing.T) {
	server := createHttpServer("localhost:9000", "testing 123", 200)
	defer func() {
		if err := server.Close(); err != nil {
			panic(err)
		}
	}()

	cleanedUp := make(chan bool)
	incoming := NewMockConn()
	defer incoming.CloseClient()
	conn := NewConnection(incoming)
	go func() {
		conn.Handle()
		cleanedUp <- true
	}()

	request := "GET http://localhost:9000/ HTTP/1.1\r\nHost: localhost\r\n\r\n"
	incoming.ClientWriter.Write([]byte(request))

	response := readMessage(incoming.ClientReader)
	expected_response := "HTTP/1.1 200 OK\r\nDate: FAKE\r\nContent-Length: 11\r\nContent-Type: text/plain; charset=utf-8\r\n\r\ntesting 123"

	if response != expected_response {
		t.Fatalf("Expected '%s' but got '%s'", expected_response, response)
	}

	incoming.CloseClient()
	<-cleanedUp
}

func TestSampleProxyWithValidAuthCredentials(t *testing.T) {
	server := createHttpServer("localhost:9000", "testing 123", 200)
	defer func() {
		if err := server.Close(); err != nil {
			panic(err)
		}
	}()

	cleanedUp := make(chan bool)
	incoming := NewMockConn()
	conn := NewConnection(incoming)
	go func() {
		conn.Handle()
		cleanedUp <- true
	}()

	setCredentials("test", "yolo")
	defer resetCredentials()
	request := "GET http://localhost:9000/ HTTP/1.1\r\nProxy-Authorization: Basic dGVzdDp5b2xv\r\nHost: localhost\r\n\r\n"
	incoming.ClientWriter.Write([]byte(request))

	response := readMessage(incoming.ClientReader)
	expected_response := "HTTP/1.1 200 OK\r\nDate: FAKE\r\nContent-Length: 11\r\nContent-Type: text/plain; charset=utf-8\r\n\r\ntesting 123"

	if response != expected_response {
		t.Fatalf("Expected '%s' but got '%s'", expected_response, response)
	}

	incoming.CloseClient()
	<-cleanedUp
}

func TestSampleProxyViaConnect(t *testing.T) {
	server := createHttpServer("localhost:9000", "testing 123", 200)
	defer func() {
		if err := server.Close(); err != nil {
			panic(err)
		}
	}()

	cleanedUp := make(chan bool)
	incoming := NewMockConn()
	conn := NewConnection(incoming)
	go func() {
		conn.Handle()
		cleanedUp <- true
	}()

	// Mimic a TLS connect here
	connect_request := "CONNECT localhost:9000 HTTP/1.1\r\nnHost: localhost:9000\r\n\r\n"
	incoming.ClientWriter.Write([]byte(connect_request))
	response := readMessage(incoming.ClientReader)
	expected_response := "HTTP/1.0 200 Connection established\r\n\r\n"
	if response != expected_response {
		t.Fatalf("Expected '%s' but got '%s'", expected_response, response)
	}

	// Mimic a regular http request
	request := "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"
	incoming.ClientWriter.Write([]byte(request))
	response = readMessage(incoming.ClientReader)
	expected_response = "HTTP/1.1 200 OK\r\nDate: FAKE\r\nContent-Length: 11\r\nContent-Type: text/plain; charset=utf-8\r\n\r\ntesting 123"
	if response != expected_response {
		t.Fatalf("Expected '%s' but got '%s'", expected_response, response)
	}

	incoming.CloseClient()
	<-cleanedUp
}

func TestParsingAddrFromHostport(t *testing.T) {
	_, err := parseAddrFromHostport("")
	if err == nil {
		t.Fatal("Expected an error.")
	}

	_, err = parseAddrFromHostport("1.1.1.1")
	if err == nil {
		t.Fatal("Expected an error.")
	}

	_, err = parseAddrFromHostport("[2001:db8::1]")
	if err == nil {
		t.Fatal("Expected an error.")
	}

	_, err = parseAddrFromHostport("somerandomstring.com")
	if err == nil {
		t.Fatal("Expected an error.")
	}

	ipv4Addr, _ := parseAddrFromHostport("1.1.1.1:8000")
	if ipv4Addr != "1.1.1.1" {
		t.Fatalf("Expected 1.1.1.1 but found %s", ipv4Addr)
	}

	ipv6Addr, _ := parseAddrFromHostport("[2001:db8::1]:80")
	if ipv6Addr != "[2001:db8::1]" {
		t.Fatalf("Expected [2001:db8::1] but found %s", ipv6Addr)
	}
}
