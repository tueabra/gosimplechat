package main

import (
	"fmt"
	"io"
	"net"
	"strings"
)

type MessageType int

const (
	NICK MessageType = 1 + iota
	JOIN
	MSG
	QUIT
)

type Message struct {
	Type     MessageType
	Client   *Client
	Contents string
}

type Server struct {
	ClientList []Client
	Incoming   chan *Message
}

func NewServer() *Server {
	s := new(Server)
	s.ClientList = make([]Client, 0)
	s.Incoming = make(chan *Message)
	return s
}

func (s *Server) ConnectionIndex(c *net.Conn) int {
	for i := range s.ClientList {
		if s.ClientList[i].Connection == *c {
			return i
		}
	}
	return -1
}

func (s *Server) Serve() {
	for {
		input := <-s.Incoming
		var msg string
		switch input.Type {
		case NICK:
			msg = fmt.Sprintf("%s is now known as %s\n", input.Contents, input.Client.Name)
		case JOIN:
			msg = fmt.Sprintf("%s has joined\n", input.Client.Name)
		case MSG:
			msg = fmt.Sprintf("%s: %s\n", input.Client.Name, input.Contents)
		case QUIT:
			msg = fmt.Sprintf("%s quit\n", input.Client.Name)
		}
		LogMessage(strings.Trim(msg, " \n\r"))
		for i := range s.ClientList {
			s.ClientList[i].Incoming <- msg
		}
	}
}

func (s *Server) HandleClient(conn net.Conn) {
	buffer := make([]byte, 1024)
	LogMessage("Incoming connection")

	newClient := NewClient(conn, *s)
	go newClient.StartMessageRelay()
	newClient.Incoming <- "Enter name: "

	bytesRead, error := conn.Read(buffer)
	if error != nil {
		if error != io.EOF {
			LogError("Client connection error: ", error)
		}
		return
	}

	name := strings.Trim(string(buffer[0:bytesRead]), " \n\r")
	newClient.Name = name

	go newClient.AcceptInput()

	s.ClientList = append(s.ClientList, *newClient)
	s.Incoming <- &Message{JOIN, newClient, ""}
}

func (s *Server) Remove(c *Client) {
	for i := range s.ClientList {
		if c.Equal(&s.ClientList[i]) {
			s.ClientList = append(s.ClientList[:i], s.ClientList[i+1:]...)
		}
	}
}
