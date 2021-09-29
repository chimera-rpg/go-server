package server

import "github.com/chimera-rpg/go-common/network"

// SendChatMessageFrom sends a message for the given client.
func (s *GameServer) SendChatMessageFrom(c *ClientConnection, original network.CommandMessage) {
	m := network.CommandMessage{
		Type: network.ChatMessage,
		From: c.GetOwner().GetTarget().Name(),
		Body: original.Body,
	}
	for _, c := range s.connectedClients {
		c.Send(m)
	}
}
