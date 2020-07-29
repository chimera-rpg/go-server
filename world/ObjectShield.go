package world

import (
	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
)

// ObjectShield represents a skill.
type ObjectShield struct {
	Object
	name        string
	damaged     float32 // How damaged the shield is.
	resistances data.AttackTypes
}

// NewObjectShield creates a skill object from the given archetype.
func NewObjectShield(a *data.Archetype) (o *ObjectShield) {
	o = &ObjectShield{
		Object: Object{Archetype: a},
	}

	//o.setArchetype(a)

	return
}

// getType returns the Archetype type.
func (o *ObjectShield) getType() cdata.ArchetypeType {
	return cdata.ArchetypeShield
}
