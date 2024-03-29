package world

import (
	"github.com/chimera-rpg/go-server/data"
)

// ObjectBlock represents walls within the game.
type ObjectBlock struct {
	Object
	//
	name  string
	maxHp int
}

// NewObjectBlock returns an ObjectBlock from the given Archetype.
func NewObjectBlock(a *data.Archetype) (o *ObjectBlock) {
	o = &ObjectBlock{
		Object: NewObject(a),
	}

	//o.setArchetype(a)

	return
}

func (o *ObjectBlock) setArchetype(targetArch *data.Archetype) {
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

func (o *ObjectBlock) getType() data.ArchetypeType {
	return data.ArchetypeBlock
}
