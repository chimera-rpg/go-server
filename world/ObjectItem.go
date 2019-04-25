package world

import (
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
		Object: Object{Archetype: *a},
	}

	// o.name, _ = a.GetValue("Name")
	if a.Name != nil {
		o.name, _ = a.Name.GetString()
	}

	return
}

func (o *ObjectItem) update(d int) {
}

func (o *ObjectItem) getType() data.ArchetypeType {
	return data.ArchetypeItem
}
