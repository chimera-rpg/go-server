package world

import (
	"github.com/chimera-rpg/go-server/data"
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
			Archetype: *a,
		},
	}

	// o.name, _ = a.GetValue("Name")
	if a.Name != nil {
		o.name, _ = a.Name.GetString()
	}

	return
}

func (o *ObjectWall) update(d int) {
}

func (o *ObjectWall) getType() data.ArchetypeType {
	return data.ArchetypeWall
}
