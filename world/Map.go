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
	mapID        data.StringID
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
		mapID: gd.MapID,
		name:  gd.Name,
	}
	gmap.owners = make([]*OwnerI, 0)
	// Size map and populate it with the data tiles
	gmap.sizeMap(gd.Height, gd.Width, gd.Depth)
	for y := range gd.Tiles {
		for x := range gd.Tiles[y] {
			for z := range gd.Tiles[y][x] {
				log.Printf("Setting %dx%dx%d\n", y, x, z)
				for a := range gd.Tiles[y][x][z] {
					object, err := gmap.CreateObjectFromArch(gm, &gd.Tiles[y][x][z][a])
					if err != nil {
						log.Print(err)
						continue
					}
					gmap.tiles[y][x][z].insertObject(object, -1)
				}
				if gmap.tiles[y][x][z].object != nil {
					target := gmap.tiles[y][x][z].object
					for ; target != nil; target = target.getNext() {
						//log.Printf("%+v\n", target)
					}
				}
				log.Print("----")
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

	// TODO: Have some sort of new owners check block for sending current Map info, initial visible tiles, and similar.

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

// CreateObjectFromArch will attempt to create an Object by an archetype, merging the result with the archetype's target Arch if possible.
func (gmap *Map) CreateObjectFromArch(gm *data.Manager, arch *data.Archetype) (o ObjectI, err error) {
	switch arch.Type {
	case data.ArchetypeFloor:
		o = ObjectI(NewObjectFloor(arch))
	case data.ArchetypeWall:
		o = ObjectI(NewObjectWall(arch))
	case data.ArchetypeItem:
		o = ObjectI(NewObjectItem(arch))
	case data.ArchetypeNPC:
		o = ObjectI(NewObjectNPC(arch))
	default:
		gameobj := ObjectGeneric{
			Object: Object{
				Archetype: arch,
			},
		}
		gameobj.value, _ = arch.Value.GetInt()
		gameobj.count, _ = arch.Count.GetInt()
		gameobj.name, _ = arch.Name.GetString()

		o = ObjectI(&gameobj)
	}

	// TODO: Create/Merge Archetype properties!
	return
}

// PlaceObject is supposed to place an object at the given x, y, and z
func (gmap *Map) PlaceObject(o ObjectI, y int, x int, z int) (err error) {
	return
}
