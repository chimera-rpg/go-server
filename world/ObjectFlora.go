package world

import (
	"github.com/chimera-rpg/go-server/data"
)

// ObjectFlora represents most non-NPC, non-PC, and non-animal living things within the game.
type ObjectFlora struct {
	Object
	//
}

// NewObjectFlora returns an ObjectFlora from the given Archetype.
func NewObjectFlora(a *data.Archetype) (o *ObjectFlora) {
	o = &ObjectFlora{
		Object: NewObject(a),
	}

	return
}

func (o *ObjectFlora) getType() data.ArchetypeType {
	return data.ArchetypeFlora
}
