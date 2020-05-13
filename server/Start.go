package server

import (
	"crypto/tls"
	"log"
	"net"
)

// Start sets up and starts handling client connections and acceptions.
func (server *GameServer) Start() (err error) {
	server.connectedClients = make(map[int]ClientConnection)
	server.clientConnections = make(chan ClientConnection)

	server.dataManager.Setup()
	server.world.Setup(&server.dataManager)

	server.listener, err = net.Listen("tcp", server.Addr)
	if err != nil {
		return err
	}
	go server.handleClientConnections()
	go server.handleClientAcceptions()
	log.Printf("Listening on %s\n", server.Addr)
	return nil
}

// SecureStart sets up and starts handling client connections and acceptions via TLS.
func (server *GameServer) SecureStart() (err error) {
	server.connectedClients = make(map[int]ClientConnection)
	server.clientConnections = make(chan ClientConnection)

	server.dataManager.Setup()
	server.world.Setup(&server.dataManager)

	cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		return err
	}
	conf := &tls.Config{Certificates: []tls.Certificate{cer}}
	server.listener, err = tls.Listen("tcp", server.Addr, conf)
	if err != nil {
		return err
	}
	go server.handleClientConnections()
	go server.handleClientAcceptions()
	log.Printf("Securely listening on %s\n", server.Addr)
	return nil
}
