package main

import (
	"io"
	"net"
	"time"
)

type MockConn struct {
	ServerReader *io.PipeReader
	ServerWriter *io.PipeWriter
	ClientReader *io.PipeReader
	ClientWriter *io.PipeWriter
}

func (c MockConn) Close() error {
	if err := c.ServerWriter.Close(); err != nil {
		return err
	}
	if err := c.ServerReader.Close(); err != nil {
		return err
	}
	return nil
}

func (c MockConn) CloseClient() error {
	if err := c.ClientWriter.Close(); err != nil {
		return err
	}
	if err := c.ClientReader.Close(); err != nil {
		return err
	}
	return nil
}

func (c MockConn) Read(data []byte) (n int, err error)  { return c.ServerReader.Read(data) }
func (c MockConn) Write(data []byte) (n int, err error) { return c.ServerWriter.Write(data) }

func (c MockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1:2342"),
		Port: 2342,
		Zone: "",
	}
}

func (c MockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1:2342"),
		Port: 2342,
		Zone: "",
	}
}

func (c MockConn) SetDeadline(t time.Time) error      { return nil }
func (c MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (c MockConn) SetWriteDeadline(t time.Time) error { return nil }

func NewMockConn() MockConn {
	serverRead, clientWrite := io.Pipe()
	clientRead, serverWrite := io.Pipe()

	return MockConn{
		ServerReader: serverRead,
		ServerWriter: serverWrite,
		ClientReader: clientRead,
		ClientWriter: clientWrite,
	}
}
