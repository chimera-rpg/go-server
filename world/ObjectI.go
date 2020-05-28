package world

import "github.com/chimera-rpg/go-server/data"

// ObjectI is the basic interface for Object access.
type ObjectI interface {
	GetOwner() OwnerI
	SetOwner(OwnerI)
	getPrevious() ObjectI
	setPrevious(ObjectI)
	getNext() ObjectI
	setNext(ObjectI)
	removeSelf()
	setArchetype(*data.Archetype)
	update(int)
	getType() data.ArchetypeType
}
