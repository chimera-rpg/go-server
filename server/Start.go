package server

import (
	"crypto/tls"
	log "github.com/sirupsen/logrus"
	"net"
	"path"
)

// Start sets up and starts handling client connections and acceptions.
func (server *GameServer) Start() (err error) {
	server.connectedClients = make(map[int]ClientConnection)
	server.clientConnections = make(chan ClientConnection)

	server.listener, err = net.Listen("tcp", server.config.Address)
	if err != nil {
		return err
	}
	go server.handleClientConnections()
	go server.handleClientAcceptions()
	log.WithFields(log.Fields{
		"Address": server.config.Address,
		"secure":  false,
	}).Print("Listening")
	return nil
}

// SecureStart sets up and starts handling client connections and acceptions via TLS.
func (server *GameServer) SecureStart() (err error) {
	server.connectedClients = make(map[int]ClientConnection)
	server.clientConnections = make(chan ClientConnection)

	serverCert := path.Join(server.dataManager.GetEtcPath(), server.config.TLSCert)
	serverKey := path.Join(server.dataManager.GetEtcPath(), server.config.TLSKey)
	cer, err := tls.LoadX509KeyPair(serverCert, serverKey)
	if err != nil {
		return err
	}
	conf := &tls.Config{Certificates: []tls.Certificate{cer}}
	server.listener, err = tls.Listen("tcp", server.config.Address, conf)
	if err != nil {
		return err
	}
	go server.handleClientConnections()
	go server.handleClientAcceptions()
	log.WithFields(log.Fields{
		"Address": server.config.Address,
		"secure":  true,
	}).Print("Securely listening")

	return nil
}
