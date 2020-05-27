package world

import (
	"github.com/chimera-rpg/go-server/data"
)

type clientConnectionI interface {
	SetOwner(p OwnerI)
	GetOwner() OwnerI
}

// OwnerPlayer represents a player character through a network
// connection and the associated player object.
type OwnerPlayer struct {
	ClientConnection clientConnectionI
	target           ObjectI
}

// getTarget returns the player's target object.
func (player *OwnerPlayer) getTarget() ObjectI {
	return player.target
}

// setTarget sets the given object as the target of the player.
func (player *OwnerPlayer) setTarget(object ObjectI) {
	player.target = object
	object.setOwner(player)
}

// NewOwnerPlayer creates a Player from a given client connection.
func NewOwnerPlayer(cc clientConnectionI, character *data.Character) *OwnerPlayer {
	return &OwnerPlayer{
		ClientConnection: cc,
		target:           NewObjectPC(character.Archetype),
	}
}

// Update does something.?
func (player *OwnerPlayer) Update(delta int64) error {
	// I guess here is where we'd have some sort of "handleCommandQueue" functionality.
	return nil
}
