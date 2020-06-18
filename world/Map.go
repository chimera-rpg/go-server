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
	owners       []OwnerI
	newOwners    []OwnerI
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
	objects      []ObjectI
	width        int
	height       int
	depth        int
}

// NewMap loads the given map file from the data manager.
func NewMap(world *World, name string) (*Map, error) {
	gm := world.data
	gd, err := gm.GetMap(name)
	if err != nil {
		return nil, fmt.Errorf("could not load map '%s'", name)
	}

	gmap := &Map{
		mapID: gd.MapID,
		name:  gd.Name,
	}
	gmap.owners = make([]OwnerI, 0)
	// Size map and populate it with the data tiles
	gmap.sizeMap(gd.Height, gd.Width, gd.Depth)
	for y := range gd.Tiles {
		for x := range gd.Tiles[y] {
			for z := range gd.Tiles[y][x] {
				log.Printf("Setting %dx%dx%d\n", y, x, z)
				for a := range gd.Tiles[y][x][z] {
					object, err := gmap.CreateObjectFromArch(world, &gd.Tiles[y][x][z][a])
					if err != nil {
						log.Print(err)
						continue
					}
					gmap.tiles[y][x][z].insertObject(object, 0)
				}
			}
		}
	}
	return gmap, nil
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

// Cleanup cleans up the given map, readying it for unloading.
func (gmap *Map) Cleanup(world *World) error {
	for y := range gmap.tiles {
		for x := range gmap.tiles[y] {
			for z := range gmap.tiles[y][x] {
				for _, o := range gmap.tiles[y][x][z].objects {
					world.objectIDs.free(o.GetID())
				}
			}
		}
	}
	return nil
}

// AddOwner adds the provided owner and its associated object to the y, x, z coordinates. This removes the owner from any previously owning maps.
func (gmap *Map) AddOwner(owner OwnerI, y, x, z int) error {
	// Remove owner from previous map.
	if m := owner.GetMap(); m != nil && m != gmap {
		m.RemoveOwner(owner)
	}

	// Set ourselves as owner's map.
	owner.SetMap(gmap)

	// Place object in our map.
	gmap.PlaceObject(owner.GetTarget(), y, x, z)

	// Send map information to owner.
	switch owner.(type) {
	case *OwnerPlayer:
		log.Println("TODO: Send OwnerPlayer our info!")
	default:
		log.Println("unhandled AddOwner")
	}

	// Add to our owners.
	gmap.owners = append(gmap.owners, owner)
	return nil
}

func (gmap *Map) RemoveOwner(owner OwnerI) error {
	if m := owner.GetMap(); m != gmap {
		return errors.New("RemoveOwner called on non-owning map")
	}

	// Remove owner's object from our map.
	if tile := owner.GetTarget().GetTile(); tile != nil {
		tile.removeObject(owner.GetTarget())
	}

	// Clear out map reference.
	owner.SetMap(nil)

	// Remove from our owners.
	for i, v := range gmap.owners {
		if v == owner {
			gmap.owners = append(gmap.owners[:i], gmap.owners[i+1:]...)
			break
		}
	}
	log.Println("Removed Owner Object")
	return nil
}

// CreateObjectFromArch will attempt to create an Object by an archetype, merging the result with the archetype's target Arch if possible.
func (gmap *Map) CreateObjectFromArch(world *World, arch *data.Archetype) (o ObjectI, err error) {
	//gm := world.data
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
				id:        world.objectIDs.acquire(),
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

// GetTile returns a pointer to the given tile.
func (gmap *Map) GetTile(y, x, z int) *Tile {
	if len(gmap.tiles) > y {
		if len(gmap.tiles[y]) > x {
			if len(gmap.tiles[y][x]) > z {
				return &gmap.tiles[y][x][z]
			}
		}
	}
	return nil
}

// PlaceObject is places an object at the given y, x, and z
func (gmap *Map) PlaceObject(o ObjectI, y int, x int, z int) (err error) {
	if o == nil {
		return errors.New("Attempted to place a nil object!")
	}
	tile := gmap.GetTile(y, x, z)
	if tile == nil {
		return errors.New("Attempted to place object out of bounds!")
	}
	tile.insertObject(o, 0)
	return
}
