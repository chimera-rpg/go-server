package world

import (
	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
)

// ObjectItem represents generic items.
type ObjectItem struct {
	Object
	//
	name  string
	maxHp int
	level int
	count int
	value int
}

// NewObjectItem creates a new ObjectItem with the passed Archetype.
func NewObjectItem(a *data.Archetype) (o *ObjectItem) {
	o = &ObjectItem{
		Object: Object{Archetype: a},
	}
	//o.setArchetype(a)

	return
}

func (o *ObjectItem) setArchetype(targetArch *data.Archetype) {
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

func (o *ObjectItem) update(d int64) {
}

func (o *ObjectItem) getType() cdata.ArchetypeType {
	return cdata.ArchetypeItem
}
