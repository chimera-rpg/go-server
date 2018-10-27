package GameWorld

import (
  "errors"
  "server/GameData"
  "log"
  "fmt"
)

type GameMap struct {
  dataName          string
  name              string
  owners            []*OwnerI
  playerCount       int
  shouldSleep       bool
  shouldExpire      bool
  lifeTime          int64 // Time in ms of how long this map has been alive
  north             *GameMap
  east              *GameMap
  south             *GameMap
  west              *GameMap
  tiles             [][]gameTile
  activeTiles       []*gameTile
  objects           []*ObjectI
  width             int
  height            int
}

func NewGameMap(gm *GameData.Manager, name string) (*GameMap, error) {
  gd, err := gm.GetMap(name)
  if err != nil {
    return nil, fmt.Errorf("Could not load map '%s'\n", name)
  }

  gmap := GameMap{
    dataName: gd.DataName,
    name: gd.Name,
  }
  gmap.owners = make([]*OwnerI, 0)
  // Size map and populate it with the data tiles
  gmap.sizeMap(gd.Width, gd.Height)
  for y := range gd.Tiles {
    for x := range gd.Tiles[y] {
      for a := range gd.Tiles[y][x] {
        object, err := gmap.CreateObjectByName(gm, gd.Tiles[y][x][a].Arch)
        if err != nil {
          continue
        }
        gmap.tiles[y][x].insertObject(object, -1)
      }
      target := gmap.tiles[y][x].object
      log.Print("----")
      for ; target != nil ; target = target.getNext() {
        log.Printf("%v\n", target)
      }
    }
  }
  return &gmap, nil
}

func (gmap *GameMap) sizeMap(width int, height int) error {
  gmap.tiles = make([][]gameTile, height)
  for y := range gmap.tiles {
    gmap.tiles[y] = make([]gameTile, width)
  }
  gmap.width = width
  gmap.height = height
  return nil
}

func (gmap *GameMap) Update(gm *GameWorld, delta int64) error {
  gmap.lifeTime += delta

  for i := range gmap.activeTiles {
    if i == 0 {
    }
  }
  /*for y := range gmap.tiles {
    for x := range gmap.tiles[y] {
    }
  }*/
  return nil
}

func (GameMap *GameMap) GetTile(x int, y int) (*gameTile, error) {
  return nil, errors.New("invalid gameTile")
}

func (gmap *GameMap) CreateObjectByName(gm *GameData.Manager, name string) (o ObjectI, err error) {
  ga, err := gm.GetArchetype(name)
  if err != nil {
    return nil, fmt.Errorf("Could not load arch '%s'\n", name)
  }

  switch ga.Type {
  case GameData.ArchetypeFloor:
    o = ObjectI(NewObjectFloor(ga))
  case GameData.ArchetypeWall:
    o = ObjectI(NewObjectWall(ga))
  case GameData.ArchetypeItem:
    o = ObjectI(NewObjectItem(ga))
  case GameData.ArchetypeNPC:
    o = ObjectI(NewObjectNPC(ga))
  default:
    gameobj := ObjectGeneric{
      Object: Object{
        Arch: name,
        Archetype: *ga,
      },
    }

    if ga.Value != nil {
      gameobj.value, _ = ga.Value.GetInt()
    }
    if ga.Count != nil {
      gameobj.count, _ = ga.Count.GetInt()
    }
    if ga.Name != nil {
      gameobj.name, _ = ga.Name.GetString()
    }

    o = ObjectI(&gameobj)
  }
  return
}

func (gm *GameMap) PlaceObject(o ObjectI, x int, y int) (err error) {
  return
}
