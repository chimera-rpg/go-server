package world

import (
	"github.com/chimera-rpg/go-server/data"
)

// ObjectFood represents a skill.
type ObjectFood struct {
	Object
	name  string
	value int32
}

// NewObjectFood creates a skill object from the given archetype.
func NewObjectFood(a *data.Archetype) (o *ObjectFood) {
	o = &ObjectFood{
		Object: Object{Archetype: a},
	}

	//o.setArchetype(a)

	return
}

// getType returns the Archetype type.
func (o *ObjectFood) getType() data.ArchetypeType {
	return data.ArchetypeFood
}
