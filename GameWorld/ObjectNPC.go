package GameWorld

import (
  "server/GameData"
)

type ObjectNPC struct {
  Object
  //
  name string
  maxHp int
  level int
  race  string
  count int
  value int
  resistance AttackTypes
  abilityScores AbilityScores
}

func NewObjectNPC(a *GameData.Archetype) (o *ObjectNPC) {
  o = &ObjectNPC{
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

func (o *ObjectNPC) update(d int) {
}

func (o *ObjectNPC) getType() GameData.ArchetypeType {
  return GameData.ArchetypeNPC
}
