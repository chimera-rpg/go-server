package server

import (
	"fmt"
	"net"
	"sync"

	"github.com/chimera-rpg/go-server/config"
	"github.com/chimera-rpg/go-server/data"
	"github.com/chimera-rpg/go-server/world"
)

// GameServer is our main server for the game. It contains the client
// connections, the world, and a data manager instance.
type GameServer struct {
	listener net.Listener
	// Client Connections
	clientConnections     chan ClientConnection
	connectedClients      map[int]ClientConnection
	connectedClientsMutex sync.Mutex
	topClientID           int
	unusedClientIDs       []int
	// Player Connections
	// players []Player.Player
	// activeMaps []Maps.Map
	world       world.World
	config      *config.Config
	dataManager data.Manager
	End         chan bool
}

// New returns a new instance of the game server.
func New() *GameServer {
	return &GameServer{
		listener: nil,
	}
}

// Setup sets up the server for use.
func (s *GameServer) Setup(cfg *config.Config) error {
	s.dataManager.Setup(cfg)
	s.world.Setup(&s.dataManager)

	// Load in our configuration
	s.config = cfg
	return nil
}

// RemoveClientByID removes a client by its ID. This comment sure added a lot.
func (s *GameServer) RemoveClientByID(id int) (err error) {
	if _, ok := s.connectedClients[id]; ok {
		delete(s.connectedClients, id)
		s.releaseClientID(id)
		return nil
	}
	return fmt.Errorf("attempted to remove non-connected ID %d", id)
}

// cleanupConnection cleans up the client, its user data, its owner data, and its object data.
func (s *GameServer) cleanupConnection(c *ClientConnection) (err error) {
	s.connectedClientsMutex.Lock()
	defer s.connectedClientsMutex.Unlock()

	// Remove object and owner data.
	fmt.Printf("cleanup data for %d\n", c.GetID())

	// Unload user data.
	if c.user != nil {
		s.dataManager.CleanupUser(c.user.Username)
	}

	// Remove the client.
	if err = s.RemoveClientByID(c.GetID()); err != nil {
		return
	}

	return
}
