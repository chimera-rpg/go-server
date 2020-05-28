package server

import (
	"fmt"
	"log"
	"net"

	"github.com/chimera-rpg/go-common/network"
	"github.com/chimera-rpg/go-server/data"
	"github.com/chimera-rpg/go-server/world"
)

// ClientConnection provides an object for storing and accessing a
// network connection.
type ClientConnection struct {
	network.Connection
	id    int
	Owner world.OwnerI
	user  *data.User
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
		if t.Type == network.Cya {
			if err := s.cleanupConnection(c); err != nil {
				log.Print(err)
			}
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
		s.cleanupConnection(c)
		c.GetSocket().Close()
		log.Print(r.(error))
		log.Print(fmt.Errorf("client %s(%d) exploded, removing", c.GetSocket().RemoteAddr().String(), c.GetID()))
	}
}

// HandleHandshake handles the client's handshake state.
func (c *ClientConnection) HandleHandshake(s *GameServer) {
	c.Send(network.Command(network.CommandHandshake{
		Version: network.Version,
		Program: "Chimera Golang Server",
	}))

	hs := c.ReceiveCommandHandshake()

	if hs.Version != network.Version {
		c.Send(network.Command(network.CommandBasic{
			Type:   network.Nokay,
			String: fmt.Sprintf("Version mismatch, expected %d, got %d", network.Version, hs.Version),
		}))
		panic(fmt.Errorf("Client version mismatch, expected %d, got %d", network.Version, hs.Version))
	}

	c.Send(network.Command(network.CommandBasic{
		Type:   network.Okay,
		String: "HAY",
	}))
	c.HandleLogin(s)
}

// HandleLogin handles the client's login state.
func (c *ClientConnection) HandleLogin(s *GameServer) {
	isWaiting := true
	var cmd network.Command
	var err error

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
			if t.Type == network.Query {
				// TODO: Query if user exists
			} else if t.Type == network.Login {
				c.user, err = s.dataManager.GetUser(t.User)
				if err != nil {
					c.Send(network.Command(network.CommandBasic{
						Type:   network.Reject,
						String: err.Error(),
					}))
				} else {
					match, err := s.dataManager.CheckUserPassword(c.user, t.Pass)
					if !match {
						c.Send(network.Command(network.CommandBasic{
							Type:   network.Reject,
							String: err.Error(),
						}))
					} else {
						c.Send(network.Command(network.CommandBasic{
							Type:   network.Okay,
							String: fmt.Sprintf("Welcome, %s!", t.User),
						}))
						isWaiting = false
					}
				}
			} else if t.Type == network.Register {
				err := s.dataManager.CreateUser(t.User, t.Pass, t.Email)
				if err != nil {
					c.Send(network.Command(network.CommandBasic{
						Type:   network.Reject,
						String: err.Error(),
					}))
				} else {
					c.Send(network.Command(network.CommandBasic{
						Type:   network.Okay,
						String: fmt.Sprintf("Hail, %s! You have been registered.", t.User),
					}))
				}
			}
		default: // Boot the client if it sends anything else.
			log.Printf("Client %s(%d) sent bad data, kicking.\n", c.GetSocket().RemoteAddr().String(), c.GetID())
			s.cleanupConnection(c)
			c.GetSocket().Close()
		}
	}

	// If we get to here, then the user has successfully logged in.
	c.HandleCharacterCreation(s)
}

// HandleCharacterCreation handles the character creation/selection of a
// connection and, potentially, sends it over to HandleGame.
func (c *ClientConnection) HandleCharacterCreation(s *GameServer) {
	isWaiting := true

	// Await an Okay response so we know the client is ready.
	for isWaiting {
		var cmd network.Command
		isHandled, shouldReturn := c.Receive(s, &cmd)
		if isHandled {
			continue
		}
		if shouldReturn {
			return
		}
		switch t := cmd.(type) {
		case network.CommandBasic:
			if t.Type == network.Okay {
				isWaiting = false
			} else {
				log.Printf("Client %s(%d) sent bad data, kicking.\n", c.GetSocket().RemoteAddr().String(), c.GetID())
				s.cleanupConnection(c)
				c.GetSocket().Close()
				return
			}
		default:
			log.Printf("Client %s(%d) sent bad data, kicking.\n", c.GetSocket().RemoteAddr().String(), c.GetID())
			s.cleanupConnection(c)
			c.GetSocket().Close()
			return
		}
	}

	//isHandled, shouldReturn := c.Receive(s, &cmd)
	// Send Genera
	/*genera := make(map[string]string)
	images := make([][]byte)
	descriptions := make([]string)
	for _, pc := range s.dataManager.GetPCArchetypes() {
		genera[pc.Properties["Genus"]] = true
		//images = append(images, s.dataManager.GetAnimImage(pc.Anim, "default", "south", 0))
		descriptions[pc.Properties["Genus"]]
	}
	c.Send(network.Command(network.CommandCharacter{
		Type:   network.QueryGenera,
		Genera: genera,
	}))
	fmt.Printf("Sending %+v\n", pc)*/

	// TODO: Send two CommandCharacter messages:
	//		* All Species, Culture, Training, Description, AbilityScores, Skills, and Images
	//		* All of the associated player's Characters as Image, Character, Level, and AbilityScores

	// Send characters.
	var playerCharacters []string
	for key := range c.user.Characters {
		playerCharacters = append(playerCharacters, key)
	}
	c.Send(network.Command(network.CommandCharacter{
		Type:       network.CreateCharacter,
		Characters: playerCharacters,
	}))

	var cmd network.Command

	isWaiting = true
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
			if t.Type == network.CreateCharacter {
				// Create a character according to species, culture, training, name
				// TODO: Maybe the Character type should have a set/array of ArchIDs to inherit from?
				// Attempt to create the character.
				if createErr := s.dataManager.CreateUserCharacter(c.user, t.Characters[0]); createErr != nil {
					c.Send(network.Command(network.CommandBasic{
						Type:   network.Reject,
						String: createErr.Error(),
					}))
					continue
				}
				// Let the client know the character exists.
				c.Send(network.Command(network.CommandCharacter{
					Type:       network.CreateCharacter,
					Characters: []string{t.Characters[0]},
				}))
			} else if t.Type == network.AdjustCharacter {
				// Changes a given character's species, culture, or training.
			} else if t.Type == network.ChooseCharacter {
				fmt.Printf("Received choose, %+v\n", t.Characters)
				if len(t.Characters) != 1 {
					// TODO: Deny request, as it is malformed.
					c.Send(network.Command(network.CommandBasic{
						Type:   network.Reject,
						String: "Invalid Characters length",
					}))
					continue
				}

				// Get the associated character.
				_, err := s.dataManager.GetUserCharacter(c.user, t.Characters[0])
				if err != nil {
					c.Send(network.Command(network.CommandBasic{
						Type:   network.Reject,
						String: err.Error(),
					}))
					continue
				}
				// TODO: We need to have the world instance handle creating a player owner and associated player object(from character). This should probably be done by either a channel or a mutex. Owner(s) in general should use channels for their communications, so that network messages can be handed over to Owners which the world can then process. Network Connection -> receives command request -> processes into an owner-compatible command -> sends it to the owner channel -> world processes it on next tick.
				// s.world.addPlayerAndCharacter(c, character) // now we can c.GetOwner() <- NewData
				// Create and set up owner corresponding to this connection.
				//c.SetOwner(world.OwnerI(world.NewOwnerPlayer(c)))
				// Create character object from its archetypes and parent it.
				//c.GetOwner().SetTarget(world.NewObjectPC(&character.Archetype))

				// Send a ChooseCharacter command to let the player know we have accepted the character.
				fmt.Println("Letting connection know the character is logging in...")
				c.Send(network.Command(network.CommandCharacter{
					Type: network.ChooseCharacter,
				}))

				isWaiting = false
				// Load a given character by name and spawn the character.
			} else if t.Type == network.DeleteCharacter {
				// Delete a given character by name.
			} else if t.Type == network.RollAbilityScores {
				// Request rolling ability scores for an in-creation character.
			}
		}
	}
	c.HandleGame(s)
}

// HandleGame handles the loop for the client when in the game state.
func (c *ClientConnection) HandleGame(s *GameServer) {
	log.Println("Now handling game connection...")
	var cmd network.Command

	for {
		isHandled, shouldReturn := c.Receive(s, &cmd)
		if isHandled {
			continue
		}
		if shouldReturn {
			return
		}
		// Handle "meta" netcommands
		// Handle
		/*switch t := cmd.(type) {
		}*/
		switch cmd.(type) {
		default: // Boot the client if it sends anything else.
			s.RemoveClientByID(c.GetID())
			c.GetSocket().Close()
			log.Printf("Client %s(%d) send bad data, kicking.\n", c.GetSocket().RemoteAddr().String(), c.GetID())
		}
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
