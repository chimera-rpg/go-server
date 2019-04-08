package GameWorld

import (
  "github.com/chimera-rpg/go-server/GameData"
)

type Object struct {
  Arch string
  GameData.Archetype
  // Relationships
  previous  ObjectI
  next      ObjectI
  parent    ObjectI
  owner     OwnerI
  //
  inventory ObjectI
}

func (object *Object) removeSelf() {
  previous := object.getPrevious()
  next := object.getNext()
  if previous != nil {
    previous.setNext(next)
  }
  if next != nil {
    next.setPrevious(previous)
  }
  object.setPrevious(nil)
  object.setNext(nil)
}

func (o *Object) update(d int) {
}

func (o *Object) getOwner() OwnerI {
  return o.owner
}
func (o *Object) setOwner(owner OwnerI) {
  // TODO: check if owner is set
  o.owner = owner
}
func (o *Object) getPrevious() ObjectI {
  return o.previous
}
func (o *Object) setPrevious(other ObjectI) {
  o.previous = other
}
func (o *Object) getNext() ObjectI {
  return o.next
}
func (o *Object) setNext(other ObjectI) {
  o.next = other
}
