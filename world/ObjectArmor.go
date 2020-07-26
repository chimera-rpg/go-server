package world

import (
	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
)

// ObjectArmor represents a skill.
type ObjectArmor struct {
	Object
	name    string
	damaged float32 // How damaged the armor is.
	armors  data.Armors
}

// NewObjectArmor creates a skill object from the given archetype.
func NewObjectArmor(a *data.Archetype) (o *ObjectArmor) {
	o = &ObjectArmor{
		Object: Object{Archetype: a},
	}

	//o.setArchetype(a)

	return
}

// getType returns the Archetype type.
func (o *ObjectArmor) getType() cdata.ArchetypeType {
	return cdata.ArchetypeArmor
}
