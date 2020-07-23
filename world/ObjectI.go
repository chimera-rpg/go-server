package world

import (
	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
	"time"
)

// ObjectI is the basic interface for Object access.
type ObjectI interface {
	GetID() ID
	GetOwner() OwnerI
	SetID(ID)
	SetOwner(OwnerI)
	SetMoved(bool)
	GetTile() *Tile
	SetTile(*Tile)
	setArchetype(*data.Archetype)
	GetArchetype() *data.Archetype
	update(time.Duration)
	getType() cdata.ArchetypeType
	AddStatus(StatusI)
	HasStatus(StatusI) bool
}
