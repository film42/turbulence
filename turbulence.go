package main

import (
	"fmt"
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
				fmt.Println("Could not accept socket:", err)
				continue
			}

			channel <- conn
		}
	}()
	return channel
}

func main() {
	fmt.Println("Prepare for takeoff...")
	server, err := net.Listen("tcp", ":25000")
	if err != nil {
		fmt.Println("Could not start server:", err)
		os.Exit(1)
	}

	fmt.Println("Server started on :25000")

	acceptedConnsChannel := acceptedConnsChannel(server)
	for {
		go handleConnection(<-acceptedConnsChannel)
	}
}
