package main

import (
	"net"
	"os"
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

	logger.Info.Println("Prepare for takeoff...")
	server, err := net.Listen("tcp", ":25000")
	if err != nil {
		logger.Fatal.Println("Could not start server:", err)
		os.Exit(1)
	}

	logger.Info.Println("Server started on :25000")

	acceptedConnsChannel := acceptedConnsChannel(server)
	for {
		go handleConnection(<-acceptedConnsChannel)
	}
}
