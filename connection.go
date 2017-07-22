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
	localAddr net.Addr
}

func (c *connection) Dial(network, address string) (net.Conn, error) {
	if !config.FollowLocalAddr {
		goto fallback
	}

	if c.localAddr == nil {
		logger.Warn.Println(c.id, "Missing local net.Addr: a default local net.Addr will be used")
		goto fallback
	}

	// Ensure the TCPAddr has its Port set to 0, which is way of telling the dialer to use
	// and random port.
	switch tcpAddr := c.localAddr.(type) {
	case *net.TCPAddr:
		tcpAddr.Port = 0
	default:
		logger.Warn.Println(c.id, "Ignoring local net.Addr", c.localAddr, "because TCPAddr was expected")
		goto fallback
	}

	dialer := &net.Dialer{LocalAddr: c.localAddr}

	// Try to dial with the incoming LocalAddr to keep the incoming and outgoing IPs the same.
	conn, err := dialer.Dial(network, address)
	if err == nil {
		return conn, nil
	}

	// If an error occurs, fallback to the default interface. This might happen if you connected
	// via a loopback interace, like testing on the same machine. We should be more specifc about
	// error handling, but falling back is fine for now.
	logger.Warn.Println(c.id, "Ignoring local net.Addr for", c.localAddr, "dialing due to error:", err)

fallback:
	return net.Dial(network, address)
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

	if !isAuthenticated(request) {
		logger.Fatal.Println(c.id, "Invalid credentials.")
		c.incoming.Write([]byte(ProxyAuthenticationRequired))
		return
	}

	// Delete the auth and proxy headers.
	if config.AuthenticationRequired() {
		request.Header.Del("Proxy-Authorization")
	}

	// Delete any other proxy related thing if enabled.
	if config.StripProxyHeaders {
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

func isAuthenticated(request *http.Request) bool {
	if !config.AuthenticationRequired() {
		return true
	}

	proxyAuthHeader := request.Header.Get("Proxy-Authorization")
	if proxyAuthHeader == "" {
		return false
	}

	username, password, ok := parseBasicAuth(proxyAuthHeader)
	if !ok {
		return false
	}

	return config.IsAuthenticated(username, password)
}

func newConnectionId() string {
	bytes := make([]byte, 3) // 6 characters long.
	if _, err := rand.Read(bytes); err != nil {
		return "[ERROR-MAKING-ID]"
	}
	return "[" + hex.EncodeToString(bytes) + "]"
}

func NewConnection(incoming net.Conn) (*connection, error) {
	newId := fmt.Sprint(newConnectionId(), " [", incoming.RemoteAddr().String(), "]")
	localAddr := incoming.LocalAddr()

	return &connection{
		id:        newId,
		incoming:  incoming,
		localAddr: localAddr,
	}, nil
}
