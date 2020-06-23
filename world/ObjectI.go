package world

import "github.com/chimera-rpg/go-server/data"

// ObjectI is the basic interface for Object access.
type ObjectI interface {
	GetID() ID
	GetOwner() OwnerI
	SetOwner(OwnerI)
	GetTile() *Tile
	SetTile(*Tile)
	setArchetype(*data.Archetype)
	GetArchetype() *data.Archetype
	update(int)
	getType() data.ArchetypeType
}
