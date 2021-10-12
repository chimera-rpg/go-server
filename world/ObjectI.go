package world

import (
	"time"

	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
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
	GetAltArchetype() *data.Archetype
	update(time.Duration)
	getType() cdata.ArchetypeType
	AddStatus(StatusI)
	RemoveStatus(StatusI) StatusI
	HasStatus(StatusI) bool
	SetStatus(StatusI) bool
	GetStatus(StatusI) StatusI
	ResolveEvent(EventI) bool
	Blocks(cdata.MatterType) bool
	Name() string
	GetDimensions() (h, w, d int)
	GetDistance(y, x, z int) float64
	//
	Stamina() int
	MaxStamina() int
	//
	RestoreStamina()
}
