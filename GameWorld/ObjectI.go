package GameWorld

import "github.com/chimera-rpg/go-server/GameData"

type ObjectI interface {
  getOwner() OwnerI
  setOwner(OwnerI)
  getPrevious() ObjectI
  setPrevious(ObjectI)
  getNext() ObjectI
  setNext(ObjectI)
  removeSelf()
  update(int)
  getType() GameData.ArchetypeType
}
