package world

import (
	"reflect"
	"time"

	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
)

// Object is the base type that should be used as an embeded struct in
// all game Objects.
type Object struct {
	Archetype *data.Archetype
	id        ID
	// Relationships
	tile   *Tile
	parent ObjectI
	owner  OwnerI
	//
	statuses  []StatusI
	inventory []ObjectI
	hasMoved  bool
	blocking  cdata.MatterType
}

// NewObject returns a new Object that references the given Archetype.
func NewObject(a *data.Archetype) Object {
	o := Object{
		blocking:  a.Blocking,
		Archetype: a,
	}
	return o
}

// update updates the given object.
func (o *Object) update(delta time.Duration) {
	for i := 0; i < len(o.statuses); i++ {
		o.statuses[i].update(delta)
		if o.statuses[i].ShouldRemove() {
			o.statuses = append(o.statuses[:i], o.statuses[i+1:]...)
			i--
		}
	}
}

// GetOwner returns the owning object.
func (o *Object) GetOwner() OwnerI {
	return o.owner
}

// SetOwner sets the owner to the given object.
func (o *Object) SetOwner(owner OwnerI) {
	// TODO: check if owner is set
	o.owner = owner
}

// SetTile sets the tile to the given tile.
func (o *Object) SetTile(tile *Tile) {
	o.tile = tile
}

// GetTile gets the tile where the object is contained.
func (o *Object) GetTile() *Tile {
	return o.tile
}

// SetMoved sets the object's hasMoved to the given value.
func (o *Object) SetMoved(b bool) {
	o.hasMoved = b
}

// SetID sets the objects' id. This should _only_ be called by World during object creation.
func (o *Object) SetID(id ID) {
	o.id = id
}

// GetID gets the object's id.
func (o *Object) GetID() ID {
	return o.id
}

func (o *Object) setArchetype(targetArch *data.Archetype) {
}

// GetArchetype gets the object's underlying archetype.
func (o *Object) GetArchetype() *data.Archetype {
	return o.Archetype
}
func (o *Object) getType() cdata.ArchetypeType {
	return cdata.ArchetypeUnknown
}

// AddStatus adds the given status to the object.
func (o *Object) AddStatus(s StatusI) {
	s.SetTarget(o)
	o.statuses = append(o.statuses, s)
	s.OnAdd()
}

// RemoveStatus removes the given status from the object.
func (o *Object) RemoveStatus(s StatusI) bool {
	for i, s2 := range o.statuses {
		if reflect.TypeOf(s) == reflect.TypeOf(s2) {
			o.statuses = append(o.statuses[:i], o.statuses[i+1:]...)
			s2.OnRemove()
			if o.GetOwner() != nil {
				o.GetOwner().SendStatus(s2, false)
			}
			return true
		}
	}
	return false
}

// HasStatus checks if the object has the given status.
func (o *Object) HasStatus(t StatusI) bool {
	for _, s := range o.statuses {
		if reflect.TypeOf(s) == reflect.TypeOf(t) {
			return true
		}
	}
	return false
}

// SetStatus sets the status.
func (o *Object) SetStatus(t StatusI) bool {
	return false
}

// ResolveEvent is the default handler for object events.
func (o *Object) ResolveEvent(e EventI) bool {
	// Do nothing per default.
	return true
}

// GetStatus returns the associated status.
func (o *Object) GetStatus(t StatusI) StatusI {
	for _, s := range o.statuses {
		if reflect.TypeOf(s) == reflect.TypeOf(t) {
			return s
		}
	}
	return nil
}

// Blocks returns if the object blocks the given MatterType.
func (o *Object) Blocks(matter cdata.MatterType) bool {
	return o.blocking.Is(matter)
}

// Name returns the name of the object, if available.
func (o *Object) Name() string {
	return ""
}

func (o *Object) GetDimensions() (h, w, d int) {
	a := o.GetArchetype()
	if a != nil {
		h = int(a.Height)
		w = int(a.Width)
		d = int(a.Depth)
	}
	if s := o.GetStatus(&StatusSqueeze{}); s != nil {
		t := s.(*StatusSqueeze)
		w -= t.X
		d -= t.Z
	}
	if s := o.GetStatus(&StatusCrouch{}); s != nil {
		t := s.(*StatusCrouch)
		h -= t.Y
	}
	return
}

// Stamina returns the object's stamina.
func (o *Object) Stamina() time.Duration {
	return 0
}

// MaxStamina returns the object's maximum stamina.
func (o *Object) MaxStamina() time.Duration {
	return 0
}
