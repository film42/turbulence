package main

import (
	"flag"
	"net"
	"os"
	"strconv"
)

var config *Config

func handleConnection(conn net.Conn) {
	connection, err := NewConnection(conn)
	if err != nil {
		logger.Fatal.Println(err)
		return
	}

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

	configPtr := flag.String("config", "", "config file")
	portPtr := flag.Int("port", 25000, "listen port")
	stripProxyHeadersPtr := flag.Bool("strip-proxy-headers", true, "strip proxy headers from http requests")
	usernamePtr := flag.String("username", "", "username for proxy authentication")
	passwordPtr := flag.String("password", "", "password for proxy authentication")
	flag.Parse()

	if *configPtr != "" {
		configFile, err := os.Open(*configPtr)
		if err != nil {
			logger.Fatal.Println("Could not open config file", err)
			os.Exit(1)
		}

		config, err = NewConfigFromReader(configFile)
		if err != nil {
			logger.Fatal.Println("Could not parse config file", err)
			os.Exit(1)
		}

		configFile.Close()
	} else {
		config = &Config{
			Port:              *portPtr,
			StripProxyHeaders: *stripProxyHeadersPtr,
		}

		if *usernamePtr != "" {
			config.Credentials = []Credential{
				{Username: *usernamePtr, Password: *passwordPtr},
			}
		}
	}

	err := config.Validate()
	if err != nil {
		logger.Fatal.Println("Config is not valid:", err)
		os.Exit(1)
	}

	if config.AuthenticationRequired() {
		logger.Info.Println("Credentials provided. Proxy authentication will be required for all connections.")
	}

	logger.Info.Println("Prepare for takeoff...")
	listenAndServe(config.Port)
}
