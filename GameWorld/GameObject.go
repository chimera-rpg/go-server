package GameWorld

import (
  "server/GameData"
  "fmt"
)

type GameObject struct {
  Arch string
  GameData.Archetype
  // Relationships
  previous  *GameObject
  next      *GameObject
  parent    *GameObject
  owner     *gameOwner
  //
  name string
  maxHp int
  level int
  race  string
  count int
  value int
  inventory *GameObject
}

func NewGameObject(gm *GameData.Manager, name string) (*GameObject, error) {
  ga, err := gm.GetArchetype(name)
  if err != nil {
    return nil, fmt.Errorf("Could not load arch '%s'\n", name)
  }

  gobject := GameObject{
    Arch: name,
    Archetype: *ga,
  }
  if ga.Value != nil {
    gobject.value, _ = ga.Value.GetInt()
  }
  if ga.Count != nil {
    gobject.count, _ = ga.Count.GetInt()
  }
  if ga.Name != nil {
    gobject.name, _ = ga.Name.GetString()
  }
  return &gobject, nil
}


func (object *GameObject) removeSelf() {
  if object.previous != nil {
    object.previous.next = object.next
  }
  if object.next != nil {
    object.next.previous = object.previous
  }
  object.previous = nil
  object.next = nil
}
