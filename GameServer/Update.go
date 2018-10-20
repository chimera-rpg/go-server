package GameServer

import (
  //"net"
  "log"
)

func (server *GameServer) Update(delta int64) error {
  go func() {
    for {
      conn, err := server.listener.Accept()
      if err != nil {
        log.Print("Error accepting: ", err.Error())
      } else {
        server.clientConnections <- *NewClientConnection(conn, server.acquireClientId())
      }
    }
  }()
  server.world.Update(delta)
  /*clients := make(chan ClientConnection)
  go generateResponses(clients)

  for {
    conn, err := net.Accept()
    if (err != nil) {
      panic(err)
    }
    Log.Print("Accepted connection.")

    go func() {
      buf := bufio.NewReader(conn)

      for {
        name, err := buf.ReadString('\n')
        if (err != nil) {
          Log.Print("Client disconnected.")
          break
        }
        clients <- ClientConnection{name, conn}
      }
    }()

  }*/
  return nil
}


func (server *GameServer) handleClientConnections() {
  for {
    clientConnection := <-server.clientConnections
    // Connected
    log.Print("New Client: ", clientConnection.GetSocket().RemoteAddr(), " as ", clientConnection.GetID())
    //
    server.connectedClients[clientConnection.GetID()] = clientConnection
    go func() {
      defer clientConnection.OnExplode(server)
      clientConnection.HandleHandshake(server)
    }()
  }
}

