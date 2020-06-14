package world

import (
	"github.com/chimera-rpg/go-server/data"
)

// MessageI is the generic interface for world messages.
type MessageI interface {
}

// MessageAddClient is a message for adding a client and its character as a Player.
type MessageAddClient struct {
	Client    clientConnectionI
	Character *data.Character
}

// MessageRemoveClient removes the player associated with the Client.
type MessageRemoveClient struct {
	Client clientConnectionI
}
