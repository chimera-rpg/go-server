package world

import (
	"github.com/chimera-rpg/go-common/network"

	"fmt"
	"log"
)

type clientConnectionI interface {
	SetOwner(p OwnerI)
	GetOwner() OwnerI
	Send(network.Command) error
	GetID() int
}

// OwnerPlayer represents a player character through a network
// connection and the associated player object.
type OwnerPlayer struct {
	commandChannel   chan OwnerCommand
	ClientConnection clientConnectionI
	target           *ObjectPC
	currentMap       *Map
}

// GetTarget returns the player's target object.
func (player *OwnerPlayer) GetTarget() ObjectI {
	return player.target
}

// SetTarget sets the given object as the target of the player.
func (player *OwnerPlayer) SetTarget(object ObjectI) {
	if objectpc, ok := object.(*ObjectPC); ok {
		player.target = objectpc
	} else {
		log.Printf("Attempted to set OwnerPlayer to non-ObjectPC...\n")
	}
	object.SetOwner(player)
}

// GetCommandChannel gets the command channel for the player.
func (player *OwnerPlayer) GetCommandChannel() chan OwnerCommand {
	return player.commandChannel
}

// SendCommand
func (player *OwnerPlayer) SendCommand(cmd network.Command) error {
	return player.ClientConnection.Send(cmd)
}

// GetMap gets the currentMap of the owner.
func (player *OwnerPlayer) GetMap() *Map {
	return player.currentMap
}

// SetMap sets the currentMap of the owner.
func (player *OwnerPlayer) SetMap(m *Map) {
	player.currentMap = m
}

// NewOwnerPlayer creates a Player from a given client connection.
func NewOwnerPlayer(cc clientConnectionI) *OwnerPlayer {
	return &OwnerPlayer{
		commandChannel:   make(chan OwnerCommand),
		ClientConnection: cc,
	}
}

// Update does something.?
func (player *OwnerPlayer) Update(delta int64) error {
	// I guess here is where we'd have some sort of "handleCommandQueue" functionality.
	done := false
	for !done {
		select {
		case pcmd, _ := <-player.commandChannel:
			fmt.Printf("Got owner command: %+v\n", pcmd)
			// Read commands
		default:
			done = true
		}
	}

	return nil
}
