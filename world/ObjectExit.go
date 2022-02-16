package world

import (
	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
)

// ObjectExit represents entrance/exit/teleporter objects.
type ObjectExit struct {
	Object
}

// NewObjectExit returns an ObjectExit from the given Archetype.
func NewObjectExit(a *data.Archetype) (o *ObjectExit) {
	o = &ObjectExit{
		Object: NewObject(a),
	}
	return
}

func (o *ObjectExit) getType() cdata.ArchetypeType {
	return cdata.ArchetypeExit
}
