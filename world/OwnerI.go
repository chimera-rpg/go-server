package world

import (
	"github.com/chimera-rpg/go-common/network"
)

// OwnerI represents the general interface that should be used
// for controlling and managing autonomous Object(s). It is used for
// Players and will eventually be used for NPCs.
type OwnerI interface {
	GetTarget() ObjectI
	SetTarget(ObjectI)
	GetCommandChannel() chan OwnerCommand
	SendCommand(network.Command) error
	SetMap(*Map)
	GetMap() *Map
	Update(delta int64) error
}
