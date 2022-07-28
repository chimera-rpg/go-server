package server

import (
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"

	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-common/network"
	"github.com/chimera-rpg/go-server/data"
	"github.com/chimera-rpg/go-server/world"
)

// ClientConnection provides an object for storing and accessing a
// network connection.
type ClientConnection struct {
	network.Connection
	id                    int
	Owner                 *world.OwnerPlayer
	user                  *data.User
	requestedAnimationIDs map[uint32]struct{}
	requestedImageIDs     map[uint32]struct{}
	requestedAudioIDs     map[uint32]struct{}
	requestedSoundIDs     map[uint32]struct{}
	log                   *log.Entry
}

// GetSocket returns the connection's socket.
func (c *ClientConnection) GetSocket() net.Conn {
	return c.Conn
}

// GetID returns the client's id.
func (c *ClientConnection) GetID() int {
	return c.id
}

// GetUser returns the client's user.
func (c *ClientConnection) GetUser() *data.User {
	return c.user
}

// NewClientConnection sets up a new ClientConnection.
func NewClientConnection(conn net.Conn, id int) *ClientConnection {
	network.RegisterCommands()
	cc := ClientConnection{
		id:                    id,
		requestedAnimationIDs: make(map[uint32]struct{}),
		requestedImageIDs:     make(map[uint32]struct{}),
		requestedAudioIDs:     make(map[uint32]struct{}),
		requestedSoundIDs:     make(map[uint32]struct{}),
	}
	cc.SetConn(conn)
	cc.log = log.WithFields(log.Fields{
		"ID":      cc.id,
		"Address": cc.GetSocket().RemoteAddr().String(),
	})
	return &cc
}

// Cleanup is called when the client connection is to be removed.
func (c *ClientConnection) Cleanup(s *GameServer) {
	s.CleanupClientChannel <- c
}

// Receive handles receiving a network command from the connection.
// It also handles error state and lower-level communications, such as
// a disconnect statement.
func (c *ClientConnection) Receive(s *GameServer, cmd *network.Command) (isHandled bool, shouldReturn bool) {
	err := c.Connection.Receive(cmd)

	if err != nil {
		panic(err)
	}

	switch t := (*cmd).(type) {
	// Here is where we'd also handle GFX requests and otherwise
	case network.CommandBasic:
		if t.Type == network.Cya {
			c.Cleanup(s)
			c.GetSocket().Close()
			c.log.Println("Client left faithfully.")
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
		c.log.Print(r.(error))
		c.log.Errorln("Client exploded, removing.")
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

	// Send Features
	c.Send(network.Command(network.CommandFeatures{
		AnimationsConfig: s.dataManager.AnimationsConfig,
	}))
	c.HandleLogin(s)
}

// HandleLogin handles the client's login state.
func (c *ClientConnection) HandleLogin(s *GameServer) {
	isWaiting := true
	var cmd network.Command
	var err error
	var goToGame bool

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
						shouldReturn = true
					} else {
						// Reconnect client to disconnected players if needed.
						owner := s.world.GetPlayerByUsername(t.User)
						if owner != nil {
							if owner.HasDummyConnection() {
								c.Owner = owner
								// Let the client know they're just reconnecting and not going to character selection.
								c.Send(network.Command(network.CommandRejoin{}))
								// Replace the world owner's connection.
								s.world.MessageChannel <- world.MessageReplaceClient{
									Client: c,
									Player: owner,
								}
								isWaiting = false
								goToGame = true
							} else {
								c.Send(network.Command(network.CommandBasic{
									Type:   network.Reject,
									String: "already connected",
								}))
								shouldReturn = true
							}
						} else { // If there is no user loaded already, then log them in normally.
							c.Send(network.Command(network.CommandBasic{
								Type:   network.Okay,
								String: fmt.Sprintf("Welcome, %s!", t.User),
							}))
							isWaiting = false
						}
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
			c.log.Warnln("Client sent bad data, kicking.")

			s.cleanupConnection(c)
			c.GetSocket().Close()
		}
	}

	// If we get to here, then the user has successfully logged in.
	if goToGame {
		c.HandleGame(s)
	} else {
		c.HandleCharacterCreation(s)
	}
}

// HandleCharacterCreation handles the character creation/selection of a
// connection and, potentially, sends it over to HandleGame.
func (c *ClientConnection) HandleCharacterCreation(s *GameServer) {
	isWaiting := true

	// Await a QueryCharacters response so we know the client is ready.
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
		case network.CommandCharacter:
			if t.Type == network.QueryCharacters {
				isWaiting = false
			} else {
				c.log.Warnln("Client sent bad data, kicking.")
				s.cleanupConnection(c)
				c.GetSocket().Close()
				return
			}
		default:
			c.log.Warnln("Client sent bad data, kicking.")
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
				character, err := s.dataManager.GetUserCharacter(c.user, t.Characters[0])
				if err != nil {
					c.Send(network.Command(network.CommandBasic{
						Type:   network.Reject,
						String: err.Error(),
					}))
					continue
				}

				// Replace the client's save info with last know Haven.
				if character.SaveInfo.HavenMap != "" {
					character.SaveInfo.Map = character.SaveInfo.HavenMap
					character.SaveInfo.Y = character.SaveInfo.HavenY
					character.SaveInfo.X = character.SaveInfo.HavenX
					character.SaveInfo.Z = character.SaveInfo.HavenZ
				}

				// Send a ChooseCharacter command to let the player know we have accepted the character.
				c.Send(network.Command(network.CommandCharacter{
					Type: network.ChooseCharacter,
				}))

				// Add the character to the world.
				s.world.MessageChannel <- world.MessageAddClient{
					Client:    c,
					Character: character,
				}

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
	var cmd network.Command

	for {
		isHandled, shouldReturn := c.Receive(s, &cmd)
		if isHandled {
			continue
		}
		if shouldReturn {
			return
		}

		switch t := cmd.(type) {
		case network.CommandAnimation:
			// If the client has already requested this animation, boot it. NOTE: It would be better to limit requests first rather than immediately booting -- as well as to warn the player that it should stop requesting.
			if _, alreadyRequested := c.requestedAnimationIDs[t.AnimationID]; alreadyRequested {
				c.log.Warnln("Kicking client due to multiple animation request")
				c.Cleanup(s)
				c.GetSocket().Close()
				return
			}
			c.requestedAnimationIDs[t.AnimationID] = struct{}{}
			if anim, err := s.dataManager.GetAnimation(t.AnimationID); err == nil {
				// This feels a bit heavy to convert our server animation data to our network animation data.
				faces := make(map[uint32][]network.AnimationFrame)
				for key, face := range anim.Faces {
					faces[key] = make([]network.AnimationFrame, len(face))
					for frameIndex, frame := range face {
						faces[key][frameIndex] = network.AnimationFrame{
							ImageID: frame.ImageID,
							Time:    frame.Time,
							X:       frame.X,
							Y:       frame.Y,
						}
					}
				}

				c.Send(network.CommandAnimation{
					AnimationID: t.AnimationID,
					RandomFrame: anim.RandomFrame,
					Faces:       faces,
				})
			} else {
				// Animation does not exist. Send client bogus data.
				c.Send(network.CommandAnimation{
					AnimationID: t.AnimationID,
				})
			}
		case network.CommandAudio:
			// If the client has already requested this audio, boot it. NOTE: It would be better to limit requests first rather than immediately booting -- as well as to warn the player that it should stop requesting.
			if _, alreadyRequested := c.requestedAudioIDs[t.AudioID]; alreadyRequested {
				c.log.Warnln("Kicking client due to multiple audio request")
				c.Cleanup(s)
				c.GetSocket().Close()
				return
			}
			c.requestedAudioIDs[t.AudioID] = struct{}{}
			if audio, err := s.dataManager.GetAudio(t.AudioID); err == nil {
				// This feels a bit heavy to convert our server audio data to our network audio data.
				sounds := make(map[uint32][]network.AudioSound)
				for key, soundSet := range audio.SoundSets {
					sounds[key] = make([]network.AudioSound, len(soundSet))
					for soundIndex, sound := range soundSet {
						sounds[key][soundIndex] = network.AudioSound{
							SoundID: sound.SoundID,
							Text:    sound.Text,
						}
					}
				}

				c.Send(network.CommandAudio{
					AudioID: t.AudioID,
					Sounds:  sounds,
				})
			} else {
				// Animation does not exist. Send client bogus data.
				c.Send(network.CommandAudio{
					AudioID: t.AudioID,
				})
			}
		case network.CommandSound:
			if _, alreadyRequested := c.requestedSoundIDs[t.SoundID]; alreadyRequested {
				c.log.Warnln("Kicking client due to multiple sound request")

				c.Cleanup(s)
				c.GetSocket().Close()
				return
			}
			c.requestedSoundIDs[t.SoundID] = struct{}{}
			if soundData, err := s.dataManager.GetSoundData(t.SoundID); err == nil {
				dataType := -1
				if string(soundData[:4]) == "fLaC" {
					dataType = network.SoundFlac
				} else if string(soundData[:4]) == "OggS" {
					dataType = network.SoundOgg
				}
				if dataType != -1 {
					c.Send(network.CommandSound{
						Type:     network.Set,
						SoundID:  t.SoundID,
						DataType: uint8(dataType),
						Data:     soundData,
					})
				}
			} else {
				// Let client know that no such graphics exists.
				c.Send(network.CommandSound{
					Type:    network.Nokay,
					SoundID: t.SoundID,
				})
			}
		case network.CommandGraphics:
			if _, alreadyRequested := c.requestedImageIDs[t.GraphicsID]; alreadyRequested {
				c.log.Warnln("Kicking client due to multiple graphics request")

				c.Cleanup(s)
				c.GetSocket().Close()
				return
			}
			c.requestedImageIDs[t.GraphicsID] = struct{}{}
			if imageData, err := s.dataManager.GetImageData(t.GraphicsID); err == nil {
				c.Send(network.CommandGraphics{
					Type:       network.Set,
					GraphicsID: t.GraphicsID,
					DataType:   network.GraphicsPng, // For now...
					Data:       imageData,
				})
			} else {
				// Let client know that no such graphics exists.
				c.Send(network.CommandGraphics{
					Type:       network.Nokay,
					GraphicsID: t.GraphicsID,
				})
			}
		case network.CommandMessage:
			switch t.Type {
			case network.ChatMessage: // General Chat
				s.SendChatMessageFrom(c, t)
			case network.MapMessage: // Local Map
			case network.PartyMessage: // Party Message
			case network.PCMessage:
				s.SendPCMessageFrom(c, t)
			default: // Bad message, boot.
				c.log.Warnln("Client sent bad data, kicking")
				c.Cleanup(s)
				c.GetSocket().Close()
			}
		case network.CommandCmd:
			c.log.WithFields(log.Fields{
				"cmd": t,
			}).Print("CommandCmd")
			switch t.Cmd {
			case network.North:
				c.Owner.GetCommandChannel() <- world.OwnerMoveCommand{Z: -1}
			case network.South:
				c.Owner.GetCommandChannel() <- world.OwnerMoveCommand{Z: 1}
			case network.East:
				c.Owner.GetCommandChannel() <- world.OwnerMoveCommand{X: 1}
			case network.West:
				c.Owner.GetCommandChannel() <- world.OwnerMoveCommand{X: -1}
			case network.Northeast:
				c.Owner.GetCommandChannel() <- world.OwnerMoveCommand{X: 1, Z: -1}
			case network.Northwest:
				c.Owner.GetCommandChannel() <- world.OwnerMoveCommand{X: -1, Z: -1}
			case network.Southeast:
				c.Owner.GetCommandChannel() <- world.OwnerMoveCommand{X: 1, Z: 1}
			case network.Southwest:
				c.Owner.GetCommandChannel() <- world.OwnerMoveCommand{X: -1, Z: 1}
			case network.Up:
				c.Owner.GetCommandChannel() <- world.OwnerMoveCommand{Y: 1}
			case network.Down:
				c.Owner.GetCommandChannel() <- world.OwnerMoveCommand{Y: -1}
			case network.Attack:
				if v, ok := t.Data.(network.CommandAttack); ok {
					c.Owner.GetCommandChannel() <- world.OwnerAttackCommand{
						Y:         int(v.Y),
						X:         int(v.X),
						Z:         int(v.Z),
						Direction: v.Direction,
						Target:    v.Target,
					}
				}
			case network.Wizard:
				if c.user.Wizard {
					c.Owner.GetCommandChannel() <- world.OwnerWizardCommand{}
				}
			}
		case network.CommandClearCmd:
			c.Owner.GetCommandChannel() <- world.OwnerClearCommand{}
		case network.CommandExtCmd:
			c.log.WithFields(log.Fields{
				"cmd": t,
			}).Print("CommandExtCmd")
			c.Owner.GetCommandChannel() <- world.OwnerExtCommand{
				Command: t.Cmd,
				Args:    t.Args,
			}
		case network.CommandRepeatCmd:
			switch t.Cmd {
			case network.North:
				c.Owner.GetCommandChannel() <- world.OwnerRepeatCommand{Command: world.OwnerMoveCommand{Z: -1}, Cancel: t.Cancel}
			case network.South:
				c.Owner.GetCommandChannel() <- world.OwnerRepeatCommand{Command: world.OwnerMoveCommand{Z: 1}, Cancel: t.Cancel}
			case network.East:
				c.Owner.GetCommandChannel() <- world.OwnerRepeatCommand{Command: world.OwnerMoveCommand{X: 1}, Cancel: t.Cancel}
			case network.West:
				c.Owner.GetCommandChannel() <- world.OwnerRepeatCommand{Command: world.OwnerMoveCommand{X: -1}, Cancel: t.Cancel}
			case network.Northeast:
				c.Owner.GetCommandChannel() <- world.OwnerRepeatCommand{Command: world.OwnerMoveCommand{X: 1, Z: -1}, Cancel: t.Cancel}
			case network.Northwest:
				c.Owner.GetCommandChannel() <- world.OwnerRepeatCommand{Command: world.OwnerMoveCommand{X: -1, Z: -1}, Cancel: t.Cancel}
			case network.Southeast:
				c.Owner.GetCommandChannel() <- world.OwnerRepeatCommand{Command: world.OwnerMoveCommand{X: 1, Z: 1}, Cancel: t.Cancel}
			case network.Southwest:
				c.Owner.GetCommandChannel() <- world.OwnerRepeatCommand{Command: world.OwnerMoveCommand{X: -1, Z: 1}, Cancel: t.Cancel}
			case network.Attack:
				if v, ok := t.Data.(network.CommandAttack); ok {
					c.Owner.GetCommandChannel() <- world.OwnerRepeatCommand{
						Command: world.OwnerAttackCommand{
							Y:         int(v.Y),
							X:         int(v.X),
							Z:         int(v.Z),
							Direction: v.Direction,
							Target:    v.Target,
						},
						Cancel: t.Cancel,
					}
				}
			}
			c.log.WithFields(log.Fields{
				"cmd": t,
			}).Print("CommandRepeatCmd")
		case network.CommandStatus:
			c.log.WithFields(log.Fields{
				"cmd": t,
			}).Print("CommandStatus")
			switch t.Type {
			case cdata.SqueezingStatus:
				c.Owner.GetCommandChannel() <- world.OwnerStatusCommand{Status: &world.StatusSqueeze{}}
			case cdata.CrouchingStatus:
				c.Owner.GetCommandChannel() <- world.OwnerStatusCommand{Status: &world.StatusCrouch{}}
			}
		case network.CommandViewport:
			// TODO: Make these limits configurable.
			if t.Height > 32 {
				t.Height = 32
			} else if t.Height < 8 {
				t.Height = 8
			}
			if t.Width > 48 {
				t.Width = 48
			} else if t.Width < 8 {
				t.Width = 8
			}
			if t.Depth > 48 {
				t.Depth = 48
			} else if t.Depth < 8 {
				t.Depth = 8
			}
			c.Owner.SetViewSize(int(t.Height), int(t.Width), int(t.Depth))
		default: // Boot the client if it sends anything else.
			c.log.Warnln("Client sent bad data, kicking")
			c.Cleanup(s)
			c.GetSocket().Close()
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
func (c *ClientConnection) GetOwner() *world.OwnerPlayer {
	return c.Owner
}

// SetOwner sets the owner(player) of this connection.
func (c *ClientConnection) SetOwner(owner *world.OwnerPlayer) {
	c.Owner = owner
}
