package GameWorld

import (
  "github.com/chimera-rpg/go-server/GameData"
)

type ObjectItem struct {
  Object
  //
  name string
  maxHp int
  level int
  count int
  value int
}

func NewObjectItem(a *GameData.Archetype) (o *ObjectItem) {
  o = &ObjectItem{
    Object: Object{Archetype: *a},
  }

  // o.name, _ = a.GetValue("Name")
  if a.Name != nil {
    o.name, _ = a.Name.GetString()
  }

  return
}

func (o *ObjectItem) update(d int) {
}

func (o *ObjectItem) getType() GameData.ArchetypeType {
  return GameData.ArchetypeItem
}
