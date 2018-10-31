package GameServer

import (
  "net"
  "log"
)

func (server *GameServer) Start() error {
  server.connectedClients = make(map[int]ClientConnection)
  server.clientConnections = make(chan ClientConnection)

  server.dataManager.Setup()
  server.world.Setup(&server.dataManager)

  var err error
  server.listener, err = net.Listen("tcp", server.Addr)
  if err != nil {
    return err
  }
  go server.handleClientConnections()
  go server.handleClientAcceptions()
  log.Printf("Listening on %s\n", server.Addr)
  return nil
}
