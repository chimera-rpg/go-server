package world

import (
	"github.com/chimera-rpg/go-server/data"
)

// Object is the base type that should be used as an embeded struct in
// all game Objects.
type Object struct {
	Archetype *data.Archetype
	// Relationships
	previous ObjectI
	next     ObjectI
	parent   ObjectI
	owner    OwnerI
	//
	inventory ObjectI
}

// removeSelf removes the given object from its linked list.
func (o *Object) removeSelf() {
	previous := o.getPrevious()
	next := o.getNext()
	if previous != nil {
		previous.setNext(next)
	}
	if next != nil {
		next.setPrevious(previous)
	}
	o.setPrevious(nil)
	o.setNext(nil)
}

// update updates the given object.
func (o *Object) update(d int) {
}

// getOwner returns the owning object.
func (o *Object) getOwner() OwnerI {
	return o.owner
}

// setOwner sets the owner to the given object.
func (o *Object) setOwner(owner OwnerI) {
	// TODO: check if owner is set
	o.owner = owner
}

// getPrevious returns the previous linked object.
func (o *Object) getPrevious() ObjectI {
	return o.previous
}

// setPrevious sets the previous linked object.
func (o *Object) setPrevious(other ObjectI) {
	o.previous = other
}

// getNext gets the next linked object.
func (o *Object) getNext() ObjectI {
	return o.next
}

// setNext sets the next linked object.
func (o *Object) setNext(other ObjectI) {
	o.next = other
}
