package world

import (
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/chimera-rpg/go-common/network"
	"github.com/chimera-rpg/go-server/data"
)

// Map is a live instance of a map that contains and updates all objects
// and tiles within it.
type Map struct {
	mapID         data.StringID
	name          string
	playerCount   int
	owners        []OwnerI
	world         *World // I guess it is okay to reference the World.
	shouldSleep   bool
	shouldExpire  bool
	lifeTime      time.Duration // Time in us of how long this map has been alive
	north         *Map
	east          *Map
	south         *Map
	west          *Map
	tiles         [][][]Tile
	activeTiles   []*Tile
	activeObjects map[ID]ObjectI
	width         int
	height        int
	depth         int
	updateTime    uint8 // Whenever this is updated, owners will check their surroundings for updates.
}

// NewMap loads the given map file from the data manager.
func NewMap(world *World, name string) (*Map, error) {
	gm := world.data
	gd, err := gm.GetMap(name)
	if err != nil {
		return nil, fmt.Errorf("could not load map '%s'", name)
	}

	gmap := &Map{
		world:         world,
		mapID:         gd.MapID,
		name:          gd.Name,
		activeObjects: make(map[ID]ObjectI),
	}
	gmap.owners = make([]OwnerI, 0)
	// Size map and populate it with the data tiles
	gmap.sizeMap(gd.Height, gd.Width, gd.Depth)
	for y := range gd.Tiles {
		for x := range gd.Tiles[y] {
			for z := range gd.Tiles[y][x] {
				for a := range gd.Tiles[y][x][z] {
					object, err := world.CreateObjectFromArch(&gd.Tiles[y][x][z][a])
					if err != nil {
						log.Warn("CreateObjectFromArch", err)
						continue
					}
					err = gmap.PlaceObject(object, y, x, z)
					if err != nil {
						log.Warn("PlaceObject", err)
					}
				}
			}
		}
	}
	return gmap, nil
}

// Stringer for dumping maps.
func (gmap *Map) String() string {
	var oIDs []uint32
	for y := range gmap.tiles {
		for x := range gmap.tiles[y] {
			for z := range gmap.tiles[y][x] {
				for _, o := range gmap.tiles[y][x][z].objects {
					oIDs = append(oIDs, o.GetID())
				}
			}
		}
	}
	return fmt.Sprintf("{name: \"%s\", height: %d, width: %d, depth: %d, owners: %d, objects: %v}", gmap.name, gmap.height, gmap.width, gmap.depth, len(gmap.owners), oIDs)
}

// sizeMap resizes the map according to the given height, width, and depth.
func (gmap *Map) sizeMap(height int, width int, depth int) error {
	gmap.tiles = make([][][]Tile, height)
	for y := range gmap.tiles {
		gmap.tiles[y] = make([][]Tile, width)
		for x := range gmap.tiles[y] {
			gmap.tiles[y][x] = make([]Tile, depth)
			for z := range gmap.tiles[y][x] {
				gmap.tiles[y][x][z] = Tile{
					gameMap: gmap,
					y:       y,
					x:       x,
					z:       z,
				}
			}
		}
	}
	gmap.width = width
	gmap.height = height
	gmap.depth = depth
	gmap.updateTime++
	return nil
}

// Update updates all active tiles and objects within the map.
func (gmap *Map) Update(gm *World, delta time.Duration) error {
	gmap.lifeTime += delta

	for _, owner := range gmap.owners {
		owner.OnMapUpdate(delta)
	}

	for _, object := range gmap.activeObjects {
		object.update(delta)
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

	if po, ok := owner.(*OwnerPlayer); ok {
		// Let client know that this object should be its view target.
		po.ClientConnection.Send(network.CommandObject{
			ObjectID: owner.GetTarget().GetID(),
			Payload:  network.CommandObjectPayloadViewTarget{},
		})
	}

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
	gmap.world.DeleteObject(owner.GetTarget(), false)

	gmap.updateTime++
	return nil
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
		return errors.New("attempted to place a nil object")
	}

	tile := gmap.GetTile(y, x, z)
	if tile == nil {
		return errors.New("attempted to place object out of bounds")
	}
	tile.insertObject(o, -1)

	tiles, _, err := gmap.GetObjectPartTiles(o, 0, 0, 0)
	for _, t := range tiles {
		t.insertObjectPart(o, -1)
	}

	// Add object types that need to update per tick.
	switch o.(type) {
	case *ObjectCharacter:
		gmap.activeObjects[o.GetID()] = o
	}

	gmap.updateTime++
	return
}

// RemoveObject removes the given object from the map.
func (gmap *Map) RemoveObject(o ObjectI) (err error) {
	if o == nil {
		return errors.New("attempted to remove a nil object")
	}

	tile := o.GetTile()
	if tile != nil {
		tile.removeObject(o)
	}

	tiles, _, err := gmap.GetObjectPartTiles(o, 0, 0, 0)
	for _, t := range tiles {
		t.removeObjectPart(o)
	}

	for _, owner := range gmap.owners {
		owner.OnObjectDelete(o.GetID())
	}

	delete(gmap.activeObjects, o.GetID())

	//gmap.updateTime++
	return
}

// MoveObject attempts to move the given object from its current position by a relative coordinate adjustment.
func (gmap *Map) MoveObject(o ObjectI, yDir, xDir, zDir int, force bool) (bool, error) {
	if o == nil {
		return false, errors.New("attempted to move a nil object")
	}

	if !force {
		var fall *StatusFalling
		if o.HasStatus(fall) {
			return false, nil
		}
	}
	// TODO: Some sort of CanMove flag, as things such as falling, paralysis, or otherwise should prevent movement. This might be handled in the calling function, such as the Owner.

	oldTiles, targetTiles, err := gmap.GetObjectPartTiles(o, yDir, xDir, zDir)
	if err != nil {
		return false, err
	}

	if len(targetTiles) == 0 {
		// Bizarre...
		return false, errors.New("somehow no tiles could be targeted")
	}

	doTilesBlock := func(targetTiles []*Tile) bool {
		matter := o.GetArchetype().Matter
		for _, tT := range targetTiles {
			for _, tO := range tT.objects {
				if tO == o {
					continue
				}
				if tO.Blocks(matter) {
					return true
				}
			}
		}
		return false
	}

	// Get our unique objects that are not this object in the target tiles.
	var uniqueObjects []ObjectI
	for _, tT := range targetTiles {
		for _, tO := range tT.objects {
			matched := false
			for _, t := range uniqueObjects {
				if t == tO {
					matched = true
					break
				}
			}
			if tO != o && !matched {
				uniqueObjects = append(uniqueObjects, tO)
			}
		}
	}

	// Get our character objects.
	var characterObjects []*ObjectCharacter
	for _, tO := range uniqueObjects {
		if t, ok := tO.(*ObjectCharacter); ok {
			characterObjects = append(characterObjects, t)
		}
	}

	// If it is blocked, check if a vertical move would solve it (if we aren't already moving vertical) -- this is for stepping up 1 unit blocks.
	if yDir == 0 {
		if doTilesBlock(targetTiles) {
			// Check if it is blocked by a character and handle that appropriately.
			if len(characterObjects) > 0 {
				log.Println("TODO: Handle character interaction")
				return false, nil
			}
			// Otherwise see if we can step down.
			_, targetUpTiles, err := gmap.GetObjectPartTiles(o, yDir+1, xDir, zDir)
			if !doTilesBlock(targetUpTiles) && err == nil {
				targetTiles = targetUpTiles
			} else {
				return false, nil
			}
		} else {
			// Check if we have to step down.
			_, targetDownTiles, err := gmap.GetObjectPartTiles(o, yDir-1, xDir, zDir)
			if !doTilesBlock(targetDownTiles) && err == nil {
				_, targetStepTiles, err := gmap.GetObjectPartTiles(o, yDir-2, xDir, zDir)
				if doTilesBlock(targetStepTiles) && err == nil {
					targetTiles = targetDownTiles
				}
			}
		}
	}
	// If we got here then the move ended up being valid, so let's update our tiles.
	// First we clear collisions from old intersection tiles.
	for _, t := range oldTiles {
		t.removeObjectPart(o)
	}
	// Second we add collisions to new intersection tiles.
	for _, t := range targetTiles {
		t.insertObjectPart(o, -1)
	}
	// Add the object to the main tile.
	targetTiles[0].insertObject(o, -1)
	gmap.updateTime++
	o.SetMoved(true)
	return true, nil
}

// GetObjectPartTiles returns two arrays for Tiles that a given object intersects with. If all directions are zero, then targetTiles will be empty.
func (gmap *Map) GetObjectPartTiles(o ObjectI, yDir, xDir, zDir int) (currentTiles, targetTiles []*Tile, err error) {
	// Get object's current root tile.
	tile := o.GetTile()
	if tile == nil {
		err = errors.New("attempted to place object out of bounds")
		return
	}
	// Get our origin.
	oY, oX, oZ := tile.y, tile.x, tile.z
	// Get our object's height, width, and depth.
	h, w, d := 1, 1, 1
	a := o.GetArchetype()
	if a != nil {
		h = int(a.Height)
		w = int(a.Width)
		d = int(a.Depth)
	}
	// Check each potential move position.
	getTargets := yDir != 0 || xDir != 0 || zDir != 0
	// Iterate through our box.
	for sY := 0; sY < h; sY++ {
		olY := oY + sY
		tY := olY + yDir
		for sX := 0; sX < w; sX++ {
			olX := oX + sX
			tX := olX + xDir
			for sZ := 0; sZ < d; sZ++ {
				olZ := oZ - sZ
				tZ := olZ + zDir
				// TODO: Only get targets as deep as the move operation!
				if getTargets {
					if tT := gmap.GetTile(tY, tX, tZ); tT != nil {
						targetTiles = append(targetTiles, tT)
					} else {
						// out of bounds.
						err = errors.New("out of bounds")
						return
					}
				}
				if oT := gmap.GetTile(olY, olX, olZ); oT != nil {
					currentTiles = append(currentTiles, oT)
				}
			}
		}
	}
	return
}
