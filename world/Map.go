package world

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"

	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
)

// Map is a live instance of a map that contains and updates all objects
// and tiles within it.
type Map struct {
	mapID        data.StringID
	name         string
	owners       []OwnerI
	world        *World // I guess it is okay to reference the World.
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
	updateTime   uint8 // Whenever this is updated, owners will check their surroundings for updates.
}

// NewMap loads the given map file from the data manager.
func NewMap(world *World, name string) (*Map, error) {
	gm := world.data
	gd, err := gm.GetMap(name)
	if err != nil {
		return nil, fmt.Errorf("could not load map '%s'", name)
	}

	gmap := &Map{
		world: world,
		mapID: gd.MapID,
		name:  gd.Name,
	}
	gmap.owners = make([]OwnerI, 0)
	// Size map and populate it with the data tiles
	gmap.sizeMap(gd.Height, gd.Width, gd.Depth)
	for y := range gd.Tiles {
		for x := range gd.Tiles[y] {
			for z := range gd.Tiles[y][x] {
				for a := range gd.Tiles[y][x][z] {
					gmap.tiles[y][x][z].y = y
					gmap.tiles[y][x][z].x = x
					gmap.tiles[y][x][z].z = z
					object, err := gmap.CreateObjectFromArch(&gd.Tiles[y][x][z][a])
					if err != nil {
						log.Print(err)
						continue
					}
					gmap.tiles[y][x][z].insertObject(object, -1)
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
	gmap.updateTime++
	return nil
}

// Update updates all active tiles and objects within the map.
func (gmap *Map) Update(gm *World, delta int64) error {
	gmap.lifeTime += delta

	for _, owner := range gmap.owners {
		owner.OnMapUpdate(delta)
	}

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

	// Add to our owners.
	gmap.owners = append(gmap.owners, owner)
	return nil
}

// RemoveOwner removes a given owner from the map.
func (gmap *Map) RemoveOwner(owner OwnerI) error {
	if m := owner.GetMap(); m != gmap {
		return errors.New("RemoveOwner called on non-owning map")
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

	// Remove object.
	gmap.DeleteObject(owner.GetTarget(), false)

	gmap.updateTime++
	return nil
}

// CreateObjectFromArch will attempt to create an Object by an archetype, merging the result with the archetype's target Arch if possible.
func (gmap *Map) CreateObjectFromArch(arch *data.Archetype) (o ObjectI, err error) {
	// Ensure archetype is compiled.
	err = gmap.world.data.CompileArchetype(arch)

	// Create our object.
	switch arch.Type {
	case cdata.ArchetypeFloor:
		o = NewObjectFloor(arch)
	case cdata.ArchetypeWall:
		o = NewObjectWall(arch)
	case cdata.ArchetypeItem:
		o = NewObjectItem(arch)
	case cdata.ArchetypeNPC:
		o = NewObjectNPC(arch)
	default:
		gameobj := ObjectGeneric{
			Object: Object{
				Archetype: arch,
				id:        gmap.world.objectIDs.acquire(),
			},
		}
		gameobj.value, _ = arch.Value.GetInt()
		gameobj.count, _ = arch.Count.GetInt()
		gameobj.name, _ = arch.Name.GetString()

		o = &gameobj
	}

	// TODO: Create/Merge Archetype properties!
	return
}

// GetTile returns a pointer to the given tile.
func (gmap *Map) GetTile(y, x, z int) *Tile {
	if len(gmap.tiles) > y && y >= 0 {
		if len(gmap.tiles[y]) > x && x >= 0 {
			if len(gmap.tiles[y][x]) > z && z >= 0 {
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
	tile.insertObject(o, -1)
	gmap.updateTime++
	return
}

// DeleteObject deletes a given object. If shouldFree is true, the associated object ID is freed.
func (gmap *Map) DeleteObject(o ObjectI, shouldFree bool) (err error) {
	if o == nil {
		return errors.New("Attempted to delete a nil object!")
	}
	if tile := o.GetTile(); tile != nil {
		tile.removeObject(o)
	}
	if shouldFree {
		gmap.world.objectIDs.free(o.GetID())
	}

	for _, owner := range gmap.owners {
		owner.OnObjectDelete(o.GetID())
	}

	return
}
