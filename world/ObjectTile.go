package world

import (
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
		Object: NewObject(a),
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

// getType returns the Archetype type.
func (o *ObjectTile) getType() data.ArchetypeType {
	return data.ArchetypeTile
}
