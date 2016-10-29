package main

import (
	"net"
	"os"
	"flag"
	"strconv"
)

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
				logger.Info.Println("Could not accept socket:", err)
				continue
			}

			channel <- conn
		}
	}()
	return channel
}

func main() {
	InitLogger()

	portPtr := flag.Int("port", 25000, "listen port")
	flag.Parse()

	logger.Info.Println("Prepare for takeoff...")

	listenOn := ":" + strconv.Itoa(*portPtr)
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
