package world

import (
	"log"

	"github.com/chimera-rpg/go-server/data"
	"github.com/imdario/mergo"
)

// ObjectWall represents walls within the game.
type ObjectWall struct {
	Object
	//
	name  string
	maxHp int
}

// NewObjectWall returns an ObjectWall from the given Archetype.
func NewObjectWall(a *data.Archetype) (o *ObjectWall) {
	o = &ObjectWall{
		Object: Object{
			Archetype: a,
		},
	}

	o.setArchetype(a)

	return
}

func (o *ObjectWall) setArchetype(targetArch *data.Archetype) {
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

func (o *ObjectWall) update(d int) {
}

func (o *ObjectWall) getType() data.ArchetypeType {
	return data.ArchetypeWall
}
