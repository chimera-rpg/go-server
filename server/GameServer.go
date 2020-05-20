package server

import (
	"fmt"
	"io/ioutil"
	"net"
	"path"

	"github.com/chimera-rpg/go-server/data"
	"github.com/chimera-rpg/go-server/world"

	"gopkg.in/yaml.v2"
)

// GameServer is our main server for the game. It contains the client
// connections, the world, and a data manager instance.
type GameServer struct {
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
	config      Config
	dataManager data.Manager
	End         chan bool
}

// New returns a new instance of the game server.
func New() *GameServer {
	return &GameServer{
		listener: nil,
	}
}

// Setup sets up the server for use, loading in configuration files and setting up data manager.
func (s *GameServer) Setup() error {
	s.dataManager.Setup()

	// Load in our configuration
	s.config = Config{
		Address: ":1337",
	}
	filepath := path.Join(s.dataManager.GetEtcPath(), "config.yml")
	r, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(r, &s.config); err != nil {
		return err
	}
	return nil
}

// RemoveClientByID removes a client by its ID. This comment sure added a lot.
func (s *GameServer) RemoveClientByID(id int) (err error) {
	if _, ok := s.connectedClients[id]; ok {
		delete(s.connectedClients, id)
		s.releaseClientID(id)
	}
	return fmt.Errorf("attempted to remove non-connected ID %d", id)
}
