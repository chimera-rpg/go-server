package GameWorld

import (
  "github.com/chimera-rpg/go-server/GameData"
)

type ObjectWall struct {
  Object
  //
  name string
  maxHp int
}

func NewObjectWall(a *GameData.Archetype) (o *ObjectWall) {
  o = &ObjectWall{
    Object: Object{
      Archetype: *a,
    },
  }

  // o.name, _ = a.GetValue("Name")
  if a.Name != nil {
    o.name, _ = a.Name.GetString()
  }

  return
}

func (o *ObjectWall) update(d int) {
}

func (o *ObjectWall) getType() GameData.ArchetypeType {
  return GameData.ArchetypeWall
}
