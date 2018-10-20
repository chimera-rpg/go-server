package GameServer

import (
  "net"
  "server/GameWorld"
  "server/GameData"
  "fmt"
)

type GameServer struct {
  Addr string
  listener net.Listener
  // Client Connections
  clientConnections chan ClientConnection
  connectedClients map[int]ClientConnection
  topClientId int
  unusedClientIds []int
  // Player Connections
  // players []Player.Player
  // activeMaps []Maps.Map
  world GameWorld.GameWorld
  dataManager GameData.Manager
  End chan bool
}

func New(addr string) *GameServer {
  return &GameServer{
    Addr: addr,
    listener: nil,
  }
}

func (s *GameServer) RemoveClientByID(id int) (err error) {
  if _, ok := s.connectedClients[id]; ok {
    delete(s.connectedClients, id)
    s.releaseClientId(id)
  }
  return fmt.Errorf("Attempted to remove non-connected ID %d\n", id)
}
