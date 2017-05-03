package main

import (
	"net/http"
	"strings"
	"testing"
)

func createHttpServer(address, payload string, code int) *http.Server {
	server := &http.Server{Addr: address}

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Date", "FAKE")
			w.WriteHeader(code)
			w.Write([]byte(payload))
		})
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	return server
}

func resetCredentials() {
	setCredentials("", "")
	AuthenticationRequired = false
}

func setCredentials(user, pass string) {
	AuthenticationRequired = true
	Username = user
	Password = pass
}

func basicHttpProxyRequest() string {
	return "GET http://httpbin.org/headers HTTP/1.1\r\nHost: httpbin.org\r\n\r\n"
}

func TestMain(m *testing.M) {
	InitLogger()
}

func TestInvalidCredentials(t *testing.T) {
	setCredentials("test", "hello")
	defer resetCredentials()

	incoming := NewMockConn()
	defer incoming.CloseClient()
	conn := NewConnection(incoming)

	go func() {
		conn.Handle()
	}()

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
	server := createHttpServer(":9000", "testing 123", 200)
	defer func() {
		if err := server.Shutdown(nil); err != nil {
			panic(err)
		}
	}()

	incoming := NewMockConn()
	defer incoming.CloseClient()
	conn := NewConnection(incoming)

	go func() {
		conn.Handle()
	}()

	request := "GET http://localhost:9000/ HTTP/1.1\r\nHost: localhost\r\n\r\n"
	incoming.ClientWriter.Write([]byte(request))

	buffer := make([]byte, 1000)
	incoming.ClientReader.Read(buffer)
	response := strings.TrimRight(string(buffer), "\x000")
	expected_response := "HTTP/1.1 200 OK\r\nDate: FAKE\r\nContent-Length: 11\r\nContent-Type: text/plain; charset=utf-8\r\n\r\ntesting 123"

	if response != expected_response {
		t.Fatalf("Expected '%s' but got '%s'", expected_response, response)
	}
}
