package world

import (
	"github.com/chimera-rpg/go-server/data"
)

// ObjectPC represents player characters.
type ObjectPC struct {
	Object
	//
	name          string
	maxHp         int
	level         int
	race          string
	count         int
	value         int
	resistance    AttackTypes
	abilityScores AbilityScores
}

// NewObjectPC creates a new ObjectPC from the given archetype.
func NewObjectPC(a *data.Archetype) (o *ObjectPC) {
	o = &ObjectPC{
		Object: Object{
			Archetype: *a,
		},
	}

	// o.name, _ = a.GetValue("Name")
	if a.Name != nil {
		o.name, _ = a.Name.GetString()
	}

	return
}

func (o *ObjectPC) update(d int) {
}

func (o *ObjectPC) getType() data.ArchetypeType {
	return data.ArchetypeNPC
}
