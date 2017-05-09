package main

import (
	"net"
	"os"
	"strconv"
	"sync"
)

type Server struct {
	waitGroup            *sync.WaitGroup
	listener             net.Listener
	acceptedConnsChannel chan net.Conn
	shuttingDown         bool
}

func NewServer() *Server {
	return &Server{
		waitGroup:            &sync.WaitGroup{},
		shuttingDown:         false,
		acceptedConnsChannel: make(chan net.Conn),
	}
}

func (server *Server) ListenAndServe(port int) {
	listenOn := ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", listenOn)
	if err != nil {
		logger.Fatal.Println("Could not start server:", err)
		os.Exit(1)
	}

	// Store reference to listener
	server.listener = listener

	logger.Info.Println("Server started on", listenOn)

	server.createListenLoop()

	for {
		go server.handleConnection(<-server.acceptedConnsChannel)
	}
}

func (server *Server) Shutdown() {
	// We send a message to ensure we never break the listener.
	server.shuttingDown = true
	server.waitGroup.Wait()
}

func (server *Server) createListenLoop() {
	go func() {
		for {
			if server.shuttingDown {
				server.listener.Close()
				return
			}

			conn, err := server.listener.Accept()
			if err != nil {
				logger.Warn.Println("Could not accept socket:", err)
				continue
			}

			server.acceptedConnsChannel <- conn
		}
	}()
}

func (server *Server) handleConnection(conn net.Conn) {
	server.waitGroup.Add(1)
	defer server.waitGroup.Done()

	connection := NewConnection(conn)
	defer connection.Close()

	connection.Handle()
}
