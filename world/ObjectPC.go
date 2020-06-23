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
	mapUpdateTime uint8 // Corresponds to the map's updateTime -- if they are out of sync then the player will sample its view space.
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

// NewObjectPCFromCharacter creates a new ObjectPC from the given character data.
func NewObjectPCFromCharacter(c *data.Character) (o *ObjectPC) {
	o = &ObjectPC{
		Object: Object{
			Archetype: &c.Archetype,
		},
		name: c.Name,
	}
	return
}

func (o *ObjectPC) setArchetype(targetArch *data.Archetype) {
	// First inherit from another Archetype if ArchID is set.
	/*mutatedArch := data.NewArchetype()
	for targetArch != nil {
		if err := mergo.Merge(&mutatedArch, targetArch); err != nil {
			log.Fatal("o no")
		}
		targetArch = targetArch.InheritArch
	}

	o.name, _ = mutatedArch.Name.GetString()*/
}

func (o *ObjectPC) update(d int) {
}

func (o *ObjectPC) getType() data.ArchetypeType {
	return data.ArchetypeNPC
}
