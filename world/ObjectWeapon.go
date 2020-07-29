package world

import (
	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
)

// ObjectWeapon represents a skill.
type ObjectWeapon struct {
	Object
	name        string
	damaged     float32 // How damaged the weapon is.
	attackTypes data.AttackTypes
	// TODO: attack types
}

// NewObjectWeapon creates a skill object from the given archetype.
func NewObjectWeapon(a *data.Archetype) (o *ObjectWeapon) {
	o = &ObjectWeapon{
		Object:      Object{Archetype: a},
		attackTypes: a.AttackTypes,
	}

	//o.setArchetype(a)

	return
}

// getType returns the Archetype type.
func (o *ObjectWeapon) getType() cdata.ArchetypeType {
	return cdata.ArchetypeWeapon
}
