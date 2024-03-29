package server

import (
	"fmt"
	"net"
	"reflect"
	"sync"

	"github.com/chimera-rpg/go-server/config"
	"github.com/chimera-rpg/go-server/data"
	"github.com/chimera-rpg/go-server/world"
	"github.com/cosmos72/gomacro/fast"
	"github.com/cosmos72/gomacro/imports"
	log "github.com/sirupsen/logrus"
)

// GameServer is our main server for the game. It contains the client
// connections, the world, and a data manager instance.
type GameServer struct {
	listener net.Listener
	// Client Connections
	clientConnections     chan ClientConnection
	CleanupClientChannel  chan *ClientConnection
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
		CleanupClientChannel: make(chan *ClientConnection),
		listener:             nil,
	}
}

// Setup sets up the server for use.
func (s *GameServer) Setup(cfg *config.Config) error {
	// Set up our interpreter globals. NOTE: All objects share the same interpreter, but use different compiled Expr for each scripting event defined. The interpreter is stored in the data package for now.
	var o world.ObjectI
	var m world.Map
	var e world.EventI
	imports.Packages["chimera"] = imports.Package{
		Binds:    map[string]reflect.Value{},
		Types:    map[string]reflect.Type{},
		Proxies:  map[string]reflect.Type{},
		Untypeds: map[string]string{},
		Wrappers: map[string][]string{},
	}

	imports.Packages["chimera"].Binds["self"] = reflect.ValueOf(&o).Elem()
	imports.Packages["chimera"].Binds["event"] = reflect.ValueOf(&e).Elem()

	data.Interpreter = fast.New()

	data.Interpreter.ImportPackage("lname", "chimera")
	data.Interpreter.ChangePackage("lname", "chimera")

	world.SetupInterpreterTypes(data.Interpreter)

	data.Interpreter.DeclVar("tile", nil, &world.Tile{})
	data.Interpreter.DeclVar("world", nil, &s.world)
	data.Interpreter.DeclVar("gamemap", nil, &m)
	data.Interpreter.DeclVar("data", nil, &s.dataManager)

	if err := s.dataManager.Setup(cfg); err != nil {
		return err
	}
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

	// Unload user data.
	if c.user != nil {
		pl := s.world.GetPlayerByUsername(c.user.Username)
		if pl == nil {
			if err = s.world.SyncPlayerSaveInfo(c); err != nil {
				log.Errorln(err)
			}
			s.dataManager.CleanupUser(c.user.Username)
		} else {
			if s.world.IsPlayerInHaven(pl) {
				if err = s.world.SyncPlayerSaveInfo(c); err != nil {
					log.Errorln(err)
				}
				s.dataManager.CleanupUser(c.user.Username)
			}
		}
	}

	// NOTE: We've adjusted the code so as to use a remove channel from client connection -> game server, so world no longer needs channel messaging since it is on the same goroutine now.
	/*s.world.MessageChannel <- world.MessageRemoveClient{
		Client: c,
	}*/
	s.world.RemovePlayerByConnection(c)

	// Remove the client.
	if err = s.RemoveClientByID(c.GetID()); err != nil {
		return
	}

	return
}

// GetConnectionIndexByUsername gets a connection by the given username if it exists.
func (s *GameServer) GetConnectionIndexByUsername(u string) int {
	for i, c := range s.connectedClients {
		if c.user.Username == u {
			return i
		}
	}
	return -1
}

// GetDataManager returns the server's data manager.
func (s *GameServer) GetDataManager() *data.Manager {
	return &s.dataManager
}

// GetWorld returns the server's world.
func (s *GameServer) GetWorld() *world.World {
	return &s.world
}
