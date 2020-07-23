package server

import (
	log "github.com/sirupsen/logrus"
	"time"
)

// Update runs update for the server world.
func (server *GameServer) Update(delta time.Duration) error {
	server.world.Update(delta)
	return nil
}

func (server *GameServer) handleClientAcceptions() {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			log.Errorln(err.Error())
		} else {
			server.connectedClientsMutex.Lock()
			clientID := server.acquireClientID()
			server.connectedClientsMutex.Unlock()
			server.clientConnections <- *NewClientConnection(conn, clientID)
		}
	}
}

func (server *GameServer) handleClientConnections() {
	for {
		clientConnection := <-server.clientConnections
		// Connected
		log.WithFields(log.Fields{
			"Address": clientConnection.GetSocket().RemoteAddr(),
			"ID":      clientConnection.GetID(),
		}).Println("New client")
		//
		server.connectedClientsMutex.Lock()
		server.connectedClients[clientConnection.GetID()] = clientConnection
		server.connectedClientsMutex.Unlock()
		go func() {
			defer clientConnection.OnExplode(server)
			clientConnection.HandleHandshake(server)
		}()
	}
}
