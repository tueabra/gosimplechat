/* Representation of a Client Connection */
package main

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"strings"
)

type Client struct {
	Name       string
	Incoming   chan string
	Connection net.Conn
	Reader     *bufio.Reader
	Writer     *bufio.Writer
	Quit       chan bool
	Server     Server
}

func NewClient(connection net.Conn, server Server) *Client {
	client := &Client{
		Name:       "",
		Incoming:   make(chan string),
		Connection: connection,
		Reader:     bufio.NewReader(connection),
		Writer:     bufio.NewWriter(connection),
		Quit:       make(chan bool),
		Server:     server,
	}
	return client
}

func (c *Client) Read() string {
	line, error := c.Reader.ReadString('\n')
	if error != nil {
		c.Close()
		if error != io.EOF {
			LogError("Client.Read() error:", error)
		}
		return ""
	}
	return line
}

func (c *Client) Close() {
	c.Quit <- true
	c.Connection.Close()
	c.Server.Remove(c)
}

func (c *Client) Equal(other *Client) bool {
	if bytes.Equal([]byte(c.Name), []byte(other.Name)) {
		if c.Connection == other.Connection {
			return true
		}
	}
	return false
}

func (c *Client) AcceptInput() {

	for {
		line := c.Read()
		if line == "" {
			break
		}
		msg := strings.Trim(string(line), " \n\r")
		if string(msg[0]) == "/" {
			if msg == "/quit" {
				c.Close()
				break
			}
			if strings.HasPrefix(msg, "/nick") {
				name := msg[6:]
				oldName := c.Name
				c.Name = name
				c.Server.Incoming <- &Message{NICK, c, oldName}
			}
		} else {
			c.Server.Incoming <- &Message{MSG, c, msg}
		}

	}

	c.Server.Incoming <- &Message{QUIT, c, ""}
}

func (c *Client) StartMessageRelay() {
	for {
		select {
		case buffer := <-c.Incoming:
			c.Writer.WriteString(buffer)
			c.Writer.Flush()
		case <-c.Quit:
			c.Connection.Close()
			break
		}
	}
}
