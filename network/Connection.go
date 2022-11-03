package network

import (
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"log"
	"net"
)

// Connection contains all needed information for network connections between clients and servers.
type Connection struct {
	IsConnected bool
	Conn        net.Conn
	Encoder     *gob.Encoder
	Decoder     *gob.Decoder
	CmdChan     chan Command  // Becomes valid for reading after ConnectTo(...). See LoopCmd
	ClosedChan  chan struct{} // Has close(...) called upon it in Close()
}

// SetConn sets the connection's net.Conn to the passed one.
func (c *Connection) SetConn(conn net.Conn) {
	if c.IsConnected == true {
		c.Close()
	}
	c.Conn = conn
	c.Encoder = gob.NewEncoder(conn)
	c.Decoder = gob.NewDecoder(conn)
	c.CmdChan = make(chan Command)
	c.ClosedChan = make(chan struct{})
	c.IsConnected = true
}

// ConnectTo connects to the given address, creating/initializing all basic fields of the Connection.
func (c *Connection) ConnectTo(address string) (err error) {
	c.Conn, err = net.Dial("tcp", address)
	if err != nil {
		return
	}
	c.Encoder = gob.NewEncoder(c.Conn)
	c.Decoder = gob.NewDecoder(c.Conn)
	c.CmdChan = make(chan Command)
	c.ClosedChan = make(chan struct{})
	c.IsConnected = true
	// I'm unsure if we should start our Command Loop channel coroutine here as it prevents and use of Send/Receive to the owner of the Connection. However, I suspect it is fine, as we should probably just use the LoopCmd 100% of the time when a client is connected to a server.
	go c.LoopCmd()
	return
}

// SecureConnectTo functions as per ConnectTo but with an additional tls.Config argument (and target TLS endpoint).
func (c *Connection) SecureConnectTo(address string, conf *tls.Config) (err error) {
	c.Conn, err = tls.Dial("tcp", address, conf)
	if err != nil {
		return
	}
	c.Encoder = gob.NewEncoder(c.Conn)
	c.Decoder = gob.NewDecoder(c.Conn)
	c.CmdChan = make(chan Command)
	c.ClosedChan = make(chan struct{})
	c.IsConnected = true
	// I'm unsure if we should start our Command Loop channel coroutine here as it prevents and use of Send/Receive to the owner of the Connection. However, I suspect it is fine, as we should probably just use the LoopCmd 100% of the time when a client is connected to a server.
	go c.LoopCmd()
	return
}

// Send sends the given Command through the connection.
func (c *Connection) Send(cmd Command) (err error) {
	err = c.Encoder.Encode(&cmd)
	return
}

// Receive a pending Command from the connection.
func (c *Connection) Receive(cmd *Command) (err error) {
	err = c.Decoder.Decode(&cmd)
	return
}

// ReceiveCommandBasic receives a basic command.
func (c *Connection) ReceiveCommandBasic() (b CommandBasic) {
	var command Command
	c.Receive(&command)
	switch t := command.(type) {
	case CommandBasic:
		b = t
	default:
		panic(fmt.Errorf("expected Net.CommandBasic(%d), got: %d", TypeBasic, t.GetType()))
	}
	return
}

// ReceiveCommandHandshake receives a handshake command.
func (c *Connection) ReceiveCommandHandshake() (hs CommandHandshake) {
	var command Command
	c.Receive(&command)
	switch t := command.(type) {
	case CommandHandshake:
		hs = t
	default:
		panic(fmt.Errorf("expected Net.CommandHandshake(%d), got: %d", TypeBasic, t.GetType()))
	}
	return
}

// Close closes a given connection. This sends a CommandBasic of Cya.
func (c *Connection) Close() {
	if c.IsConnected == false {
		return
	}
	c.IsConnected = false
	if r := recover(); r != nil {
		log.Print("Closing due to problematic connection.")
	} else {
		c.Send(CommandBasic{
			Type: Cya,
		})
	}
	c.Conn.Close()
	var blank struct{}
	c.ClosedChan <- blank
}

// LoopCmd is a loop that receives commands and pumps them into the CmdChan.
func (c *Connection) LoopCmd() {
	var cmd Command
	var err error
	for c.IsConnected {
		err = c.Receive(&cmd)
		if err != nil {
			c.Close()
			break
		}
		c.CmdChan <- cmd
	}
}
