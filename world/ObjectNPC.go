package world

import (
	"github.com/chimera-rpg/go-server/data"
)

// ObjectNPC represents non player characters.
type ObjectNPC struct {
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

// NewObjectNPC creates a new ObjectNPC from the given archetype.
func NewObjectNPC(a *data.Archetype) (o *ObjectNPC) {
	o = &ObjectNPC{
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

func (o *ObjectNPC) update(d int) {
}

func (o *ObjectNPC) getType() data.ArchetypeType {
	return data.ArchetypeNPC
}
