package GameWorld

import (
  "github.com/chimera-rpg/go-server/GameData"
)

type ObjectGeneric struct {
  Object
  //
  name string
  maxHp int
  level int
  race  string
  count int
  value int
}

func (o *ObjectGeneric) update(d int) {
}

func (o *ObjectGeneric) getType() GameData.ArchetypeType {
  return GameData.ArchetypeGeneric
}
