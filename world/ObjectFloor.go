package world

import (
	"github.com/chimera-rpg/go-server/data"
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
		Object: Object{Archetype: *a},
	}

	// o.name, _ = a.GetValue("Name")
	if a.Name != nil {
		o.name, _ = a.Name.GetString()
	}

	return
}

// update updates the floor.
func (o *ObjectFloor) update(d int) {
}

// getType returns the Archetype type.
func (o *ObjectFloor) getType() data.ArchetypeType {
	return data.ArchetypeFloor
}
