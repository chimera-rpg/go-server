package server

import (
	"log"
)

// Update runs update for the server world.
func (server *GameServer) Update(delta int64) error {
	server.world.Update(delta)
	return nil
}

func (server *GameServer) handleClientAcceptions() {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			log.Print("Error accepting: ", err.Error())
		} else {
			server.clientConnections <- *NewClientConnection(conn, server.acquireClientID())
		}
	}
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
