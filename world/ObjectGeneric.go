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

// NewObjectGeneric creates a new ObjectGeneric with the passed Archetype.
func NewObjectGeneric(a *data.Archetype) (o *ObjectGeneric) {
	o = &ObjectGeneric{
		Object: Object{Archetype: a},
	}
	//o.setArchetype(a)
	return
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

func (o *ObjectGeneric) getType() cdata.ArchetypeType {
	return cdata.ArchetypeGeneric
}

// SetName sets the generic object's name.
func (o *ObjectGeneric) SetName(name string) {
	o.name = name
}

// Name returns the generic object's name.
func (o *ObjectGeneric) Name() string {
	return o.name
}

// GetMundaneInfo returns the mundane info of the object.
func (o *ObjectGeneric) GetMundaneInfo(near bool) cdata.ObjectInfo {
	info := cdata.ObjectInfo{
		Name: o.Name(),
	}
	if near {
		info.Matter = o.Matter()
	}
	return info
}
