package world

import (
	"github.com/chimera-rpg/go-server/data"
)

// ObjectGeneric represents generic objects that have no defined type.
type ObjectGeneric struct {
	Object
	//
	name  string
	maxHp int
	level int
	race  string
	count int
	value int
}

func (o *ObjectGeneric) update(d int) {
}

func (o *ObjectGeneric) getType() data.ArchetypeType {
	return data.ArchetypeGeneric
}
