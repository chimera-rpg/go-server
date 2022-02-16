package world

import (
	"errors"
	"fmt"
	"time"

	"github.com/cosmos72/gomacro/fast"
	log "github.com/sirupsen/logrus"

	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-common/network"
	"github.com/chimera-rpg/go-server/data"
)

// Map is a live instance of a map that contains and updates all objects
// and tiles within it.
type Map struct {
	mapID          data.StringID
	name           string
	dataName       string
	playerCount    int
	owners         []OwnerI
	world          *World // I guess it is okay to reference the World.
	shouldSleep    bool
	shouldExpire   bool
	lifeTime       time.Duration // Time in us of how long this map has been alive
	north          *Map
	east           *Map
	south          *Map
	west           *Map
	tiles          [][][]Tile
	activeTiles    []*Tile
	activeObjects  map[ID]ObjectI
	width          int
	height         int
	depth          int
	y, x, z        int           // Default entry point.
	updateTime     uint8         // Whenever this is updated, owners will check their surroundings for updates.
	turnTime       time.Duration // Time until the next map turn (when characters have their actions restored)
	turnElapsed    time.Duration
	refreshObjects []ID
	interpreter    *fast.Interp
	handlers       MapHandlers
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
		dataName:      gd.DataName,
		activeObjects: make(map[ID]ObjectI),
		y:             gd.Y,
		x:             gd.X,
		z:             gd.Z,
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
					object.ResolveEvent(EventBirth{})
					if err != nil {
						log.Warn("PlaceObject", err)
					}
				}
			}
		}
	}

	// Add interpreter as needed.
	if gd.Script != "" {
		gmap.addInterpreter(gd.Script)
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
					Y:       y,
					X:       x,
					Z:       z,
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

	// Force owners to forget about objects in the refreshObjects list.
	if len(gmap.refreshObjects) > 0 {
		for _, oID := range gmap.refreshObjects {
			for _, owner := range gmap.owners {
				owner.ForgetObject(oID)
			}
		}
		gmap.refreshObjects = make([]uint32, 0)
	}

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

	// This might be a bit heavy...
	if gmap.handlers.updateFunc != nil {
		gmap.handlers.updateFunc(delta)
	}

	/*for y := range gmap.tiles {
	  for x := range gmap.tiles[y] {
	  }
	}*/
	return nil
}

// Cleanup cleans up the given map, readying it for unloading.
func (gmap *Map) Cleanup(world *World) error {
	if gmap.handlers.cleanupFunc != nil {
		gmap.handlers.cleanupFunc()
	}

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

	// Finally, call scripting.
	if gmap.handlers.ownerJoinFunc != nil {
		gmap.handlers.ownerJoinFunc(owner)
	}

	return nil
}

// RemoveOwner removes a given owner from the map.
func (gmap *Map) RemoveOwner(owner OwnerI) error {
	if m := owner.GetMap(); m != gmap {
		return errors.New("RemoveOwner called on non-owning map")
	}

	// Finally, call scripting.
	if gmap.handlers.ownerLeaveFunc != nil {
		gmap.handlers.ownerLeaveFunc(owner)
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

	tiles, _, err := gmap.GetObjectPartTiles(o, 0, 0, 0, false)
	for _, t := range tiles {
		t.insertObjectPart(o, -1)
	}

	// Add object types that need to update per tick.
	switch o.(type) {
	case *ObjectCharacter:
		gmap.activeObjects[o.GetID()] = o
	case *ObjectAudio:
		gmap.activeObjects[o.GetID()] = o
	default:
		if o.Updates() {
			gmap.activeObjects[o.GetID()] = o
		}
		// If the object has timers, add it.
		if len(*o.Timers()) > 0 {
			gmap.activeObjects[o.GetID()] = o
		}
	}

	gmap.updateTime++
	return
}

// RemoveObject removes the given object from the map.
func (gmap *Map) RemoveObject(o ObjectI) (err error) {
	if o == nil {
		return errors.New("attempted to remove a nil object")
	}

	tiles, _, err := gmap.GetObjectPartTiles(o, 0, 0, 0, false)
	for _, t := range tiles {
		t.removeObjectPart(o)
	}

	tile := o.GetTile()
	if tile != nil {
		tile.removeObject(o)
	}

	for _, owner := range gmap.owners {
		owner.OnObjectDelete(o.GetID())
	}

	delete(gmap.activeObjects, o.GetID())

	//gmap.updateTime++
	return
}

// TeleportObject teleports the given object from its current position to an absolute position.
func (gmap *Map) TeleportObject(o ObjectI, y, x, z int, force bool) error {
	yDir := y - o.GetTile().Y
	xDir := x - o.GetTile().X
	zDir := z - o.GetTile().Z

	oldTiles, targetTiles, err := gmap.GetObjectPartTiles(o, yDir, xDir, zDir, force)
	if err != nil {
		return err
	}

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

	return nil
}

// MoveObject attempts to move the given object from its current position by a relative coordinate adjustment.
func (gmap *Map) MoveObject(o ObjectI, yDir, xDir, zDir int, force bool) (bool, error) {
	if o == nil {
		return false, errors.New("attempted to move a nil object")
	}

	if !force {
		if o.HasStatus(StatusFallingRef) {
			return false, nil
		}
	}
	// TODO: Some sort of CanMove flag, as things such as falling, paralysis, or otherwise should prevent movement. This might be handled in the calling function, such as the Owner.

	oldTiles, targetTiles, err := gmap.GetObjectPartTiles(o, yDir, xDir, zDir, true)
	if err != nil {
		return false, err
	}

	if len(targetTiles) == 0 {
		// Bizarre...
		return false, errors.New("somehow no tiles could be targeted")
	}

	// Check if we're uncrouching or should be crouched.
	if crouch := o.GetStatus(StatusCrouchRef); crouch != nil {
		s := crouch.(*StatusCrouch)
		if s.Remove {
			o.RemoveStatus(crouch)
			_, targetTiles, err = gmap.GetObjectPartTiles(o, yDir, xDir, zDir, true)
			if err != nil {
				return false, err
			}
			if DoTilesBlock(o, targetTiles) {
				s.Remove = false
				o.AddStatus(s)
				o.GetOwner().SendMessage("There is not enough space to stand here!") // TODO: Replace with an Event or something.
				h, w, d := o.GetDimensions()
				audioID := gmap.world.data.Strings.Acquire("bonk")
				soundID := gmap.world.data.Strings.Acquire("default")
				gmap.EmitSound(audioID, soundID, targetTiles[0].Y+h, targetTiles[0].X+w/2, targetTiles[0].Z+d/2, 0.25)
				return false, nil
			}
		} else if !s.Crouching {
			_, targetTiles, err = gmap.GetObjectPartTiles(o, yDir, xDir, zDir, true)
			if err != nil {
				return false, err
			}
			s.Crouching = true
		}
	}

	// Check if we're unsqueezing or should be squeezed.
	if squeeze := o.GetStatus(StatusSqueezeRef); squeeze != nil {
		s := squeeze.(*StatusSqueeze)
		if s.Remove {
			o.RemoveStatus(squeeze)
			_, targetTiles, err = gmap.GetObjectPartTiles(o, yDir, xDir, zDir, true)
			if err != nil {
				return false, err
			}
			if DoTilesBlock(o, targetTiles) {
				s.Remove = false
				o.AddStatus(s)
				o.GetOwner().SendMessage("There is not enough space to unsqueeze here!") // TODO: Replace with an Event or something.
				return false, nil
			}
		} else if !s.Squeezing {
			_, targetTiles, err = gmap.GetObjectPartTiles(o, yDir, xDir, zDir, true)
			if err != nil {
				return false, err
			}
			s.Squeezing = true
		}
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

	// If we will intersect with a touch exit, the object should be teleported(?).
	// TODO: Move to ObjectCharacter.update?
	if _, ok := o.(*ObjectCharacter); ok {
		for _, tO := range uniqueObjects {
			if t, ok := tO.(*ObjectExit); ok {
				if t.Archetype.Exit != nil && t.Archetype.Exit.Touch {
					if err := t.Teleport(o); err == nil {
						return true, nil
					} else {
						log.Printf("Couldn't use touch teleporter: %s\n", err)
					}
				}
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
		if DoTilesBlock(o, targetTiles) {
			// Check if it is blocked by a character and handle that appropriately.
			if len(characterObjects) > 0 {
				log.Println("TODO: Handle character interaction")
				return false, nil
			}
			// Otherwise see if we can step down.
			_, targetUpTiles, err := gmap.GetObjectPartTiles(o, yDir+1, xDir, zDir, false)
			if !DoTilesBlock(o, targetUpTiles) && err == nil {
				targetTiles = targetUpTiles
			} else {
				return false, nil
			}
		} else {
			// Only attempt to move down if we're not flying or floating.
			if !o.HasStatus(StatusFlyingRef) && !o.HasStatus(StatusFloatingRef) {
				// Check if we have to step down.
				_, targetDownTiles, err := gmap.GetObjectPartTiles(o, yDir-1, xDir, zDir, false)
				if !DoTilesBlock(o, targetDownTiles) && err == nil {
					_, targetStepTiles, err := gmap.GetObjectPartTiles(o, yDir-2, xDir, zDir, false)
					if DoTilesBlock(o, targetStepTiles) && err == nil {
						targetTiles = targetDownTiles
					}
				}
			}
		}
	} else if !force {
		// If we're requesting to move up/down, only allow it if the object swims or flies.
		if !o.HasStatus(StatusFlyingRef) && !o.HasStatus(StatusSwimmingRef) {
			return false, nil
		}
		if DoTilesBlock(o, targetTiles) {
			return false, nil
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
func (gmap *Map) GetObjectPartTiles(o ObjectI, yDir, xDir, zDir int, force bool) (currentTiles, targetTiles []*Tile, err error) {
	// Get object's current root tile.
	tile := o.GetTile()
	if tile == nil {
		err = errors.New("attempted to place object out of bounds")
		return
	}
	// Get our origin.
	oY, oX, oZ := tile.Y, tile.X, tile.Z
	// Get our object's height, width, and depth.
	h, w, d := o.GetDimensions()
	// Check each potential move position.
	getTargets := force || yDir != 0 || xDir != 0 || zDir != 0

	// FIXME: We're treating tile types as having their part(s) one Y step below their actual position. This is so that tile types provide collision that matches their visual position. This feels Wrong(tm).
	if o.getType() == cdata.ArchetypeTile {
		oY--
	}

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

// RefreshObject marks the given object to be refreshed.
func (gmap *Map) RefreshObject(oID ID) {
	if o, ok := gmap.world.objects[oID]; ok {
		o.GetTile().modTime++
		gmap.refreshObjects = append(gmap.refreshObjects, oID)
		gmap.updateTime++
	}
}

// InactivateObject removes the given object from the active objects map.
func (gmap *Map) InactiveObject(oID ID) {
	delete(gmap.activeObjects, oID)
}

// ActivateObject adds the given object to the active objects map.
func (gmap *Map) ActivateObject(oID ID) {
	if o, ok := gmap.world.objects[oID]; ok {
		gmap.activeObjects[oID] = o
	}
}

// Sounds

// EmitSound emits a sound at Y, X, Z to all characters at a volume.
func (gmap *Map) EmitSound(audioID, soundID ID, y, x, z int, volume float32) {
	for _, o := range gmap.activeObjects {
		switch c := o.(type) {
		case *ObjectCharacter:
			c.HandleSound(audioID, soundID, y, x, z, volume)
		}
	}
}

// EmitObjectSound emits a sound at an object to all characters at a volume.
func (gmap *Map) EmitObjectSound(audioID, soundID ID, o ObjectI, volume float32) {
	for _, o := range gmap.activeObjects {
		switch c := o.(type) {
		case *ObjectCharacter:
			c.HandleObjectSound(audioID, soundID, o, volume)
		}
	}
}

// TODO: Move these helper functions

func DoTilesBlock(o ObjectI, targetTiles []*Tile) bool {
	matter := o.Matter()
	for _, tT := range targetTiles {
		for _, tO := range tT.objectParts {
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

func IsInLiquid(targetTiles []*Tile) bool {
	liquidThreshold := len(targetTiles) - len(targetTiles)/3
	liquidCount := 0
	for _, tT := range targetTiles {
		if tT.matter.Is(cdata.LiquidMatter) {
			liquidCount++
		}
		if liquidCount >= liquidThreshold {
			return true
		}
	}
	return false
}
