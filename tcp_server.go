package tcp_server

import (
	"bufio"
	"crypto/tls"
	"log"
	"net"
)

// Client holds info about connection
type Client struct {
	conn   net.Conn
	server *Server
}

// TCP Server
type Server struct {
	address                  string // Address to open connection: localhost:9999
	config                   *tls.Config
	onNewClientCallback      func(c *Client)
	onClientConnectionClosed func(c *Client, err error)
	onNewMessage             func(c *Client, message string)
}

// Read client data from channel
func (c *Client) listen() {
	c.server.onNewClientCallback(c)
	reader := bufio.NewReader(c.conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			c.conn.Close()
			c.server.onClientConnectionClosed(c, err)
			return
		}
		c.server.onNewMessage(c, message)
	}
}

// Send text message to client
func (c *Client) Send(message string) error {
	_, err := c.conn.Write([]byte(message))
	return err
}

// Send bytes to client
func (c *Client) SendBytes(b []byte) error {
	_, err := c.conn.Write(b)
	return err
}

func (c *Client) Conn() net.Conn {
	return c.conn
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// Called right after Server starts listening new client
func (s *Server) OnNewClient(callback func(c *Client)) {
	s.onNewClientCallback = callback
}

// Called right after connection closed
func (s *Server) OnClientConnectionClosed(callback func(c *Client, err error)) {
	s.onClientConnectionClosed = callback
}

// Called when Client receives new message
func (s *Server) OnNewMessage(callback func(c *Client, message string)) {
	s.onNewMessage = callback
}

// Listen starts network Server
func (s *Server) Listen() {
	var listener net.Listener
	var err error
	if s.config == nil {
		listener, err = net.Listen("tcp", s.address)
	} else {
		listener, err = tls.Listen("tcp", s.address, s.config)
	}
	if err != nil {
		log.Fatal("Error starting TCP Server.")
	}
	defer listener.Close()

	for {
		conn, _ := listener.Accept()
		client := &Client{
			conn:   conn,
			server: s,
		}
		go client.listen()
	}
}

// Creates new tcp Server instance
func New(address string) *Server {
	log.Println("Creating Server with address", address)
	Server := &Server{
		address: address,
		config:  nil,
	}

	Server.OnNewClient(func(c *Client) {})
	Server.OnNewMessage(func(c *Client, message string) {})
	Server.OnClientConnectionClosed(func(c *Client, err error) {})

	return Server
}

func NewWithTLS(address string, certFile string, keyFile string) *Server {
	log.Println("Creating Server with address", address)
	cert, _ := tls.LoadX509KeyPair(certFile, keyFile)
	config := tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	Server := &Server{
		address: address,
		config:  &config,
	}

	Server.OnNewClient(func(c *Client) {})
	Server.OnNewMessage(func(c *Client, message string) {})
	Server.OnClientConnectionClosed(func(c *Client, err error) {})

	return Server
}
