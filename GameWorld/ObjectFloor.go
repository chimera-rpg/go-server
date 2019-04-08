package GameWorld

import (
  "github.com/chimera-rpg/go-server/GameData"
)

type ObjectFloor struct {
  Object
  name string
  slow int
}

func NewObjectFloor(a *GameData.Archetype) (o *ObjectFloor) {
  o = &ObjectFloor{
    Object: Object{Archetype: *a},
  }

  // o.name, _ = a.GetValue("Name")
  if a.Name != nil {
    o.name, _ = a.Name.GetString()
  }

  return
}

func (o *ObjectFloor) update(d int) {
}

func (o *ObjectFloor) getType() GameData.ArchetypeType {
  return GameData.ArchetypeFloor
}
