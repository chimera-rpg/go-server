package world

import (
	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
)

// ObjectTile represents a tile type object.
type ObjectTile struct {
	Object
	name string
	slow int
}

// NewObjectTile creates a floor object from the given archetype.
func NewObjectTile(a *data.Archetype) (o *ObjectTile) {
	o = &ObjectTile{
		Object: Object{Archetype: a},
	}

	//o.setArchetype(a)

	return
}

func (o *ObjectTile) setArchetype(targetArch *data.Archetype) {
	// First inherit from another Archetype if ArchID is set.
	/*baseArch := data.NewArchetype()
	for targetArch != nil {
		if err := mergo.Merge(&baseArch, targetArch); err != nil {
			log.Fatal("o no")
		}
		targetArch = targetArch.InheritArch
	}

	o.name, _ = targetArch.Name.GetString()*/
}

// update updates the floor.
func (o *ObjectTile) update(d int64) {
}

// getType returns the Archetype type.
func (o *ObjectTile) getType() cdata.ArchetypeType {
	return cdata.ArchetypeTile
}
