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

// SendPCMessageFrom sends a "say" message for the given client.
func (s *GameServer) SendPCMessageFrom(c *ClientConnection, original network.CommandMessage) {
	m := network.CommandMessage{
		Type:         network.PCMessage,
		From:         c.GetOwner().GetTarget().Name(),
		FromObjectID: c.GetOwner().GetTarget().GetID(),
		Body:         original.Body,
	}
	// TODO: Get characters within a radius of PC!
	for _, c := range s.connectedClients {
		c.Send(m)
	}
}
