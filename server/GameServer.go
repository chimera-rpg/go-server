package server

import (
	"fmt"
	"net"

	"github.com/chimera-rpg/go-server/data"
	"github.com/chimera-rpg/go-server/world"
)

// GameServer is our main server for the game. It contains the client
// connections, the world, and a data manager instance.
type GameServer struct {
	Addr     string
	listener net.Listener
	// Client Connections
	clientConnections chan ClientConnection
	connectedClients  map[int]ClientConnection
	topClientID       int
	unusedClientIDs   []int
	// Player Connections
	// players []Player.Player
	// activeMaps []Maps.Map
	world       world.World
	dataManager data.Manager
	End         chan bool
}

// New returns a new instance of the game server.
func New(addr string) *GameServer {
	return &GameServer{
		Addr:     addr,
		listener: nil,
	}
}

// RemoveClientByID removes a client by its ID. This comment sure added a lot.
func (s *GameServer) RemoveClientByID(id int) (err error) {
	if _, ok := s.connectedClients[id]; ok {
		delete(s.connectedClients, id)
		s.releaseClientID(id)
	}
	return fmt.Errorf("attempted to remove non-connected ID %d", id)
}
