package world

import (
	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
)

// ObjectGeneric represents generic objects that have no defined type.
type ObjectGeneric struct {
	Object
	//
	name  string
	maxHp int
	level int
	race  string
	count int
	value int
}

func (o *ObjectGeneric) setArchetype(targetArch *data.Archetype) {
	// First inherit from another Archetype if ArchID is set.
	/*baseArch := data.NewArchetype()
	for targetArch != nil {
		if err := mergo.Merge(&baseArch, targetArch); err != nil {
			log.Fatal("o no")
		}
		targetArch = targetArch.InheritArch
	}*/
}

func (o *ObjectGeneric) update(d int) {
}

func (o *ObjectGeneric) getType() cdata.ArchetypeType {
	return cdata.ArchetypeGeneric
}
