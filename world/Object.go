package world

import (
	"github.com/chimera-rpg/go-server/data"
)

// Some object states...
const (
	FallingState int64 = 1 << iota
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
	inventory ObjectI
	state     int64
	hasMoved  bool
}

// update updates the given object.
func (o *Object) update(d int64) {
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

// SetMoved
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

// GetArchetype gets the object's underlying archetype.
func (o *Object) GetArchetype() *data.Archetype {
	return o.Archetype
}
