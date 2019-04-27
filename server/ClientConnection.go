package server

import (
	"fmt"
	"log"
	"net"

	"github.com/chimera-rpg/go-common/network"
	"github.com/chimera-rpg/go-server/world"
)

// ClientConnection provides an object for storing and accessing a
// network connection.
type ClientConnection struct {
	network.Connection
	id    int
	Owner world.OwnerI
}

// GetSocket returns the connection's socket.
func (c *ClientConnection) GetSocket() net.Conn {
	return c.Conn
}

// GetID returns the client's id.
func (c *ClientConnection) GetID() int {
	return c.id
}

// NewClientConnection sets up a new ClientConnection.
func NewClientConnection(conn net.Conn, id int) *ClientConnection {
	network.RegisterCommands()
	cc := ClientConnection{
		id: id,
	}
	cc.SetConn(conn)
	return &cc
}

// Receive handles receiving a network command from the connection.
// It also handles error state and lower-level communications, such as
// a disconnect statement.
func (c *ClientConnection) Receive(s *GameServer, cmd *network.Command) (isHandled bool, shouldReturn bool) {
	err := c.Connection.Receive(cmd)

	if err != nil {
		panic(fmt.Errorf("client %s(%d) exploded, removing", c.GetSocket().RemoteAddr().String(), c.GetID()))
	}

	switch t := (*cmd).(type) {
	// Here is where we'd also handle GFX requests and otherwise
	case network.CommandBasic:
		if t.Type == network.CYA {
			s.RemoveClientByID(c.GetID())
			c.GetSocket().Close()
			log.Printf("Client %s(%d) left faithfully.\n", c.GetSocket().RemoteAddr().String(), c.GetID())
			isHandled = true
			shouldReturn = true
		}
		isHandled = false
	}
	return
}

// OnExplode handles when the client explodes.
func (c *ClientConnection) OnExplode(s *GameServer) {
	if r := recover(); r != nil {
		s.RemoveClientByID(c.GetID())
		c.GetSocket().Close()
		log.Print(r.(error))
		log.Print(fmt.Errorf("client %s(%d) exploded, removing", c.GetSocket().RemoteAddr().String(), c.GetID()))
	}
}

// HandleHandshake handles the client's handshake state.
func (c *ClientConnection) HandleHandshake(s *GameServer) {
	c.Send(network.Command(network.CommandHandshake{
		Version: network.VERSION,
		Program: "Chimera Golang Server",
	}))

	hs := c.ReceiveCommandHandshake()

	if hs.Version != network.VERSION {
		c.Send(network.Command(network.CommandBasic{
			Type:   network.NOK,
			String: fmt.Sprintf("Version mismatch, expected %d, got %d", network.VERSION, hs.Version),
		}))
		panic(fmt.Errorf("Client version mismatch, expected %d, got %d", network.VERSION, hs.Version))
	}

	c.Send(network.Command(network.CommandBasic{
		Type:   network.OK,
		String: "HAY",
	}))
	c.HandleLogin(s)
}

// HandleLogin handles the client's login state.
func (c *ClientConnection) HandleLogin(s *GameServer) {
	isWaiting := true
	var cmd network.Command

	for isWaiting {
		isHandled, shouldReturn := c.Receive(s, &cmd)
		if isHandled {
			continue
		}
		if shouldReturn {
			return
		}
		switch t := cmd.(type) {
		case network.CommandLogin:
			if t.Type == network.QUERY {
				// TODO: Query if user exists
			} else if t.Type == network.LOGIN {
				user, err := s.dataManager.GetUser(t.User)
				if err != nil {
					c.Send(network.Command(network.CommandBasic{
						Type:   network.REJECT,
						String: err.Error(),
					}))
				}
				if user.Password != t.Pass {
					c.Send(network.Command(network.CommandBasic{
						Type:   network.REJECT,
						String: "bad password",
					}))
				} else {
					c.Send(network.Command(network.CommandBasic{
						Type:   network.OK,
						String: "Welcome :)",
					}))
					isWaiting = false
				}
			} else if t.Type == network.REGISTER {
				// FIXME: we're not handling err in the case of access problems
				user, _ := s.dataManager.GetUser(t.User)
				if user != nil {
					c.Send(network.Command(network.CommandBasic{
						Type:   network.REJECT,
						String: "user exists",
					}))
					continue
				}
				//	user, err := s.dataManager.CreateUser(t.User, t.Pass, t.Email)
				c.Send(network.Command(network.CommandBasic{
					Type:   network.OK,
					String: "Welcome, new user :)",
				}))
				isWaiting = false
			}
		default: // Boot the client if it sends anything else.
			s.RemoveClientByID(c.GetID())
			c.GetSocket().Close()
			log.Printf("Client %s(%d) send bad data, kicking.\n", c.GetSocket().RemoteAddr().String(), c.GetID())
		}
	}
	// If we get to here, then the user has successfully logged in.
	c.HandleCharacterCreation(s)
}

// HandleCharacterCreation handles the character creation/selection of a
// connection and, potentially, sends it over to HandleGame.
func (c *ClientConnection) HandleCharacterCreation(s *GameServer) {
	isWaiting := true

	var cmd network.Command
	for isWaiting {
		isHandled, shouldReturn := c.Receive(s, &cmd)
		if shouldReturn {
			return
		}
		if isHandled {
			continue
		}
		switch t := cmd.(type) {
		case network.CommandCharacter:
			if t.Type == network.QUERY_RACE {
				// Return results for given race by name.
			} else if t.Type == network.QUERY_CLASS {
				// Return results for given class by name.
			} else if t.Type == network.QUERY_CHARACTER {
				// Return full description of existing character.
			} else if t.Type == network.CREATE_CHARACTER {
				// Create a character according to race,class,name
			} else if t.Type == network.LOAD_CHARACTER {
				// Load a given character by name and spawn the character.
			} else if t.Type == network.DELETE_CHARACTER {
				// Delete a given character by name.
			} else if t.Type == network.ROLL_ABILITY_SCORES {
				// Request rolling ability scores for an in-creation character.
			}
		}
	}
	//c.HandleGame(s)
}

// HandleGame handles the loop for the client when in the game state.
func (c *ClientConnection) HandleGame(s *GameServer) {
	var cmd network.Command

	for {
		isHandled, shouldReturn := c.Receive(s, &cmd)
		if isHandled {
			continue
		}
		if shouldReturn {
			return
		}
		/*switch t := cmd.(type) {
		}*/
	}

}

// HandleTravel handles the state of a client traveling into a map.
func (c *ClientConnection) HandleTravel(s *GameServer, m *world.Map) {
	//var cmd network.Command
	// Get list of unique archetype images in the map
	// Send <CRC32>->PNG data for each
}

// GetOwner returns the owner(player) of this connection.
func (c *ClientConnection) GetOwner() world.OwnerI {
	return c.Owner
}

// SetOwner sets the owner(player) of this connection.
func (c *ClientConnection) SetOwner(owner world.OwnerI) {
	c.Owner = owner
}
