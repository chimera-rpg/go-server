package world

import (
	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
)

// ObjectSpecial represents special objects for in-map features.
type ObjectSpecial struct {
	Object
}

// NewObjectSpecial creates the special object.
func NewObjectSpecial(a *data.Archetype) (o *ObjectSpecial) {
	o = &ObjectSpecial{
		Object: NewObject(a),
	}

	return
}

// getType returns the Archetype type.
func (o *ObjectSpecial) getType() cdata.ArchetypeType {
	return cdata.ArchetypeSpecial
}
