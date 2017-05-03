package main

import (
	"flag"
	"net"
	"os"
	"strconv"
)

var AuthenticationRequired = false
var Username = ""
var Password = ""
var StripProxyHeaders = true

func handleConnection(conn net.Conn) {
	connection := NewConnection(conn)
	defer connection.Close()
	connection.Handle()
}

func acceptedConnsChannel(listener net.Listener) chan net.Conn {
	channel := make(chan net.Conn)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				logger.Warn.Println("Could not accept socket:", err)
				continue
			}

			channel <- conn
		}
	}()
	return channel
}

func validCredentials(username, password string) bool {
	if username == "" && password == "" {
		return true
	}
	if username != "" && password != "" {
		AuthenticationRequired = true
		return true
	}
	return false
}

func listenAndServe(port int) {
	listenOn := ":" + strconv.Itoa(port)
	server, err := net.Listen("tcp", listenOn)
	if err != nil {
		logger.Fatal.Println("Could not start server:", err)
		os.Exit(1)
	}

	logger.Info.Println("Server started on", listenOn)

	acceptedConnsChannel := acceptedConnsChannel(server)
	for {
		go handleConnection(<-acceptedConnsChannel)
	}
}

func main() {
	InitLogger()

	portPtr := flag.Int("port", 25000, "listen port")
	usernamePtr := flag.String("username", "", "username for proxy authentication")
	passwordPtr := flag.String("password", "", "password for proxy authentication")
	stripProxyHeadersPtr := flag.Bool("strip-proxy-headers", true, "strip proxy headers from http requests")
	flag.Parse()

	if !validCredentials(*usernamePtr, *passwordPtr) {
		logger.Fatal.Println("Invalid credentials provided. Must have a username/password or none at all.")
		os.Exit(1)
	}

	if AuthenticationRequired {
		logger.Info.Println("Credentials provided. Proxy authentication will be required for all connections.")
		Username = *usernamePtr
		Password = *passwordPtr
	}

	StripProxyHeaders = *stripProxyHeadersPtr

	logger.Info.Println("Prepare for takeoff...")
	listenAndServe(*portPtr)
}
