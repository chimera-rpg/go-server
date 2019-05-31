package world

import (
	"errors"
	"fmt"
	"log"

	"github.com/chimera-rpg/go-server/data"
)

// Map is a live instance of a map that contains and updates all objects
// and tiles within it.
type Map struct {
	dataName     string
	name         string
	owners       []*OwnerI
	playerCount  int
	shouldSleep  bool
	shouldExpire bool
	lifeTime     int64 // Time in ms of how long this map has been alive
	north        *Map
	east         *Map
	south        *Map
	west         *Map
	tiles        [][][]Tile
	activeTiles  []*Tile
	objects      []*ObjectI
	width        int
	height       int
	depth        int
}

// NewMap loads the given map file from the data manager.
func NewMap(gm *data.Manager, name string) (*Map, error) {
	gd, err := gm.GetMap(name)
	if err != nil {
		return nil, fmt.Errorf("could not load map '%s'", name)
	}

	gmap := Map{
		dataName: gd.DataName,
		name:     gd.Name,
	}
	gmap.owners = make([]*OwnerI, 0)
	// Size map and populate it with the data tiles
	gmap.sizeMap(gd.Height, gd.Width, gd.Depth)
	for y := range gd.Tiles {
		for x := range gd.Tiles[y] {
			for z := range gd.Tiles[y][x] {
				log.Printf("Setting %dx%dx%d\n", y, x, z)
				for a := range gd.Tiles[y][x][z] {
					object, err := gmap.CreateObjectByArchID(gm, gd.Tiles[y][x][z][a].ArchID)
					if err != nil {
						continue
					}
					gmap.tiles[y][x][z].insertObject(object, -1)
				}
				target := gmap.tiles[y][x][z].object
				log.Print("----")
				for ; target != nil; target = target.getNext() {
					//log.Printf("%+v\n", target)
				}
			}
		}
	}
	return &gmap, nil
}

// sizeMaps resizes the map according to the given height, width, and depth.
func (gmap *Map) sizeMap(height int, width int, depth int) error {
	gmap.tiles = make([][][]Tile, height)
	for y := range gmap.tiles {
		gmap.tiles[y] = make([][]Tile, width)
		for x := range gmap.tiles[y] {
			gmap.tiles[y][x] = make([]Tile, depth)
		}

	}
	gmap.width = width
	gmap.height = height
	gmap.depth = depth
	return nil
}

// Update updates all active tiles and objects within the map.
func (gmap *Map) Update(gm *World, delta int64) error {
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

// GetTile returns the tile at the given x, y, and z.
func (gmap *Map) GetTile(y int, x int, z int) (*Tile, error) {
	return nil, errors.New("invalid Tile")
}

// CreateObjectByArchID will attempt to create an Object by its archetype id.
func (gmap *Map) CreateObjectByArchID(gm *data.Manager, id data.FileID) (o ObjectI, err error) {
	ga, err := gm.GetArchetype(id)
	if err != nil {
		return nil, fmt.Errorf("could not load arch '%d'", id)
	}

	switch ga.Type {
	case data.ArchetypeFloor:
		o = ObjectI(NewObjectFloor(ga))
	case data.ArchetypeWall:
		o = ObjectI(NewObjectWall(ga))
	case data.ArchetypeItem:
		o = ObjectI(NewObjectItem(ga))
	case data.ArchetypeNPC:
		o = ObjectI(NewObjectNPC(ga))
	default:
		gameobj := ObjectGeneric{
			Object: Object{
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

// PlaceObject is supposed to place an object at the given x, y, and z
func (gmap *Map) PlaceObject(o ObjectI, y int, x int, z int) (err error) {
	return
}
