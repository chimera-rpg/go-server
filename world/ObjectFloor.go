package world

import (
	"log"

	"github.com/chimera-rpg/go-server/data"
	"github.com/imdario/mergo"
)

// ObjectFloor represents a floor type object.
type ObjectFloor struct {
	Object
	name string
	slow int
}

// NewObjectFloor creates a floor object from the given archetype.
func NewObjectFloor(a *data.Archetype) (o *ObjectFloor) {
	o = &ObjectFloor{
		Object: Object{Archetype: a},
	}

	o.setArchetype(a)

	return
}

func (o *ObjectFloor) setArchetype(targetArch *data.Archetype) {
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

// update updates the floor.
func (o *ObjectFloor) update(d int) {
}

// getType returns the Archetype type.
func (o *ObjectFloor) getType() data.ArchetypeType {
	return data.ArchetypeFloor
}
