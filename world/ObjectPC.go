package world

import (
	"log"

	"github.com/chimera-rpg/go-server/data"
	"github.com/imdario/mergo"
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
			Archetype: a,
		},
	}

	//o.setArchetype(a)

	return
}

func (o *ObjectPC) setArchetype(targetArch *data.Archetype) {
	// First inherit from another Archetype if ArchID is set.
	mutatedArch := data.NewArchetype()
	for targetArch != nil {
		if err := mergo.Merge(&mutatedArch, targetArch); err != nil {
			log.Fatal("o no")
		}
		targetArch = targetArch.InheritArch
	}

	o.name, _ = mutatedArch.Name.GetString()
}

func (o *ObjectPC) update(d int) {
}

func (o *ObjectPC) getType() data.ArchetypeType {
	return data.ArchetypeNPC
}
