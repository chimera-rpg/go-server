package world

import "github.com/chimera-rpg/go-server/data"

// ObjectI is the basic interface for Object access.
type ObjectI interface {
	getOwner() OwnerI
	setOwner(OwnerI)
	getPrevious() ObjectI
	setPrevious(ObjectI)
	getNext() ObjectI
	setNext(ObjectI)
	removeSelf()
	update(int)
	getType() data.ArchetypeType
}
