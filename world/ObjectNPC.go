package world

import (
	"log"

	"github.com/chimera-rpg/go-server/data"
	"github.com/imdario/mergo"
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
			Archetype: a,
		},
	}

	//o.setArchetype(a)

	return
}

func (o *ObjectNPC) setArchetype(targetArch *data.Archetype) {
	// First inherit from another Archetype if ArchID is set.
	baseArch := data.NewArchetype()
	for targetArch != nil {
		if err := mergo.Merge(&baseArch, targetArch); err != nil {
			log.Fatal("o no")
		}
		targetArch = targetArch.InheritArch
	}

	o.name, _ = targetArch.Name.GetString()
}

func (o *ObjectNPC) update(d int) {
}

func (o *ObjectNPC) getType() data.ArchetypeType {
	return data.ArchetypeNPC
}
