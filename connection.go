package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

const ProxyAuthenticationRequired = "HTTP/1.0 407 Proxy authentication required\r\n\r\n"

type connection struct {
	id       string
	incoming net.Conn
	outgoing net.Conn
	proxy
}

func (c *connection) Handle() {
	logger.Info.Println(c.id, "Handling new connection.")

	reader := bufio.NewReader(c.incoming)
	request, err := http.ReadRequest(reader)
	if err == io.EOF {
		logger.Warn.Println(c.id, "Incoming connection disconnected.")
		return
	}
	if err != nil {
		logger.Warn.Println(c.id, "Could not parse or read request from incoming connection:", err)
		return
	}

	defer request.Body.Close()

	// TODO: Make this return a 407
	if AuthenticationRequired && !credentialsAreValid(request) {
		logger.Fatal.Println(c.id, "Invalid credentials.")
		c.incoming.Write([]byte(ProxyAuthenticationRequired))
		return
	}

	// Delete the auth and proxy headers.
	if AuthenticationRequired {
		request.Header.Del("Proxy-Authorization")
	}

	// Delete any other proxy related thing if enabled.
	if StripProxyHeaders {
		request.Header.Del("Forwarded")
		request.Header.Del("Proxy-Connection")
		request.Header.Del("Via")
		request.Header.Del("X-Forwarded-For")
		request.Header.Del("X-Forwarded-Host")
		request.Header.Del("X-Forwarded-Proto")
	}

	logger.Info.Println(c.id, "Processing connection to:", request.Method, request.Host)

	if request.Method == "CONNECT" {
		c.proxy = &httpsProxy{}
	} else {
		c.proxy = &httpProxy{}
	}

	err = c.proxy.SetupOutgoing(c, request)
	if err != nil {
		logger.Warn.Println(c.id, err)
		return
	}

	// Spawn incoming->outgoing and outgoing->incoming streams.
	signal := make(chan error, 1)
	go streamBytes(c.incoming, c.outgoing, signal)
	go streamBytes(c.outgoing, c.incoming, signal)

	// Wait for either stream to complete and finish. The second will always be an error.
	err = <-signal
	if err != nil {
		logger.Warn.Println(c.id, "Error reading or writing data", request.Host, err)
		return
	}
}

func (c *connection) Close() {
	if c.incoming != nil {
		c.incoming.Close()
	}

	if c.outgoing != nil {
		c.outgoing.Close()
	}

	logger.Info.Println(c.id, "Connection closed.")
}

// COPIED FROM STD LIB TO USE WITH PROXY-AUTH HEADER
// parseBasicAuth parses an HTTP Basic Authentication string.
// "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==" returns ("Aladdin", "open sesame", true).
func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}

func credentialsAreValid(request *http.Request) bool {
	proxyAuthHeader := request.Header.Get("Proxy-Authorization")
	if proxyAuthHeader == "" {
		return false
	}

	username, password, ok := parseBasicAuth(proxyAuthHeader)

	if !ok {
		return false
	}

	if username == Username && password == Password {
		return true
	}

	return false
}

func newConnectionId() string {
	bytes := make([]byte, 3) // 6 characters long.
	if _, err := rand.Read(bytes); err != nil {
		return "[ERROR-MAKING-ID]"
	}
	return "[" + hex.EncodeToString(bytes) + "]"
}

func NewConnection(incoming net.Conn) *connection {
	newId := fmt.Sprint(newConnectionId(), " [", incoming.RemoteAddr().String(), "]")

	return &connection{
		id:       newId,
		incoming: incoming,
	}
}
