package GameWorld

import (
  "errors"
  "server/GameData"
  "log"
  "fmt"
)

type gameMap struct {
  dataName          string
  name              string
  owners            []*gameOwner
  playerCount       int
  shouldSleep       bool
  shouldExpire      bool
  lifeTime          int64 // Time in ms of how long this map has been alive
  north             *gameMap
  east              *gameMap
  south             *gameMap
  west              *gameMap
  tiles             [][]gameTile
  activeTiles       []*gameTile
  objects           []*GameObject
  width             int
  height            int
}

func NewGameMap(gm *GameData.Manager, name string) (*gameMap, error) {
  gd, err := gm.GetMap(name)
  if err != nil {
    return nil, fmt.Errorf("Could not load map '%s'\n", name)
  }

  gmap := gameMap{
    dataName: gd.DataName,
    name: gd.Name,
  }
  gmap.owners = make([]*gameOwner, 0)
  // Size map and populate it with the data tiles
  gmap.sizeMap(gd.Width, gd.Height)
  for y := range gd.Tiles {
    for x := range gd.Tiles[y] {
      for a := range gd.Tiles[y][x] {
        object, err := NewGameObject(gm, gd.Tiles[y][x][a].Arch)
        if err != nil {
          continue
        }
        gmap.tiles[y][x].insertObject(object, -1)
      }
      target := gmap.tiles[y][x].object
      log.Print("----")
      for ; target != nil ; target = target.next {
        log.Printf("%v\n", target)
      }
    }
  }
  return &gmap, nil
}

func (gmap *gameMap) sizeMap(width int, height int) error {
  gmap.tiles = make([][]gameTile, height)
  for y := range gmap.tiles {
    gmap.tiles[y] = make([]gameTile, width)
  }
  gmap.width = width
  gmap.height = height
  return nil
}

func (gmap *gameMap) Update(gm *GameWorld, delta int64) error {
  gmap.lifeTime += delta
  /*
  for i := range gmap.activeTiles {
    log.Print("Updating activeTile ", i)
  }*/
  /*for y := range gmap.tiles {
    for x := range gmap.tiles[y] {
    }
  }*/
  return nil
}

func (gameMap *gameMap) GetTile(x int, y int) (*gameTile, error) {
  return nil, errors.New("invalid gameTile")
}
