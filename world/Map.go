package world

import (
	"errors"
	"fmt"
	"math"
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
	lightObjects   map[ID]ObjectI
	actions        []ActionI // Actions that are added and processed each update.
	width          int
	height         int
	depth          int
	y, x, z        int           // Default entry point.
	haven          bool          // If the map is a haven.
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
		lightObjects:  make(map[ID]ObjectI),
		y:             gd.Y,
		x:             gd.X,
		z:             gd.Z,
		haven:         gd.Haven,
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

	// Generate map tile's sky lighting logic.
	gmap.RefreshSky()

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
					gameMap:      gmap,
					Y:            y,
					X:            x,
					Z:            z,
					lightModTime: 1, // Set to 1 to ensure difference from player view's default light mod time.
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

	// Refresh our actions.
	gmap.actions = make([]ActionI, 0)

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

	// Process our actions.
	for _, action := range gmap.actions {
		if action.Ready() {
			switch a := action.(type) {
			case *ActionMove:
				if _, err := a.object.GetTile().GetMap().MoveObject(a.object, a.y, a.x, a.z, false); err != nil {
					log.Warn(err)
				}
			case *ActionAttack:
				if a.Target != 0 {
					o2 := gmap.world.GetObject(a.Target)
					if o2 == nil {
						log.Errorln("Attack request for missing object")
						continue
					}
					if o2.GetTile().GetMap() != gmap {
						log.Errorln("Attack request for object in different map")
						continue
					}
					if o2.Attackable() {
						switch attacker := a.object.(type) {
						case *ObjectCharacter:
							if attacker.Attack(o2) {
								break
							}
						}
					}
				} else if a.Y != 0 || a.X != 0 || a.Z != 0 {
					h, w, d := a.object.GetDimensions()
					t := a.object.GetTile()
					tiles := gmap.ShootRay(float64(t.Y)+float64(h)/2, float64(t.X)+float64(w)/2, float64(t.Z)+float64(d)/2, float64(a.Y), float64(a.X), float64(a.Z), func(t *Tile) bool {
						return true
					})
					objs := getUniqueObjectsInTiles(tiles)
					// Ignore our own tile.
					for _, o := range objs {
						// Ignore ourself.
						if o == a.object {
							continue
						}
						if o.Attackable() {
							switch attacker := a.object.(type) {
							case *ObjectCharacter:
								if attacker.Attack(o) {
									break
								}
							}
						}
					}
				}
			case *ActionStatus:
				a.object.SetStatus(a.status)
			case *ActionSpawn:
				HandleActionSpawn(gmap, a)
			}
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
			Payload: network.CommandObjectPayloadViewTarget{
				Height: uint8(po.viewHeight),
				Width:  uint8(po.viewWidth),
				Depth:  uint8(po.viewDepth),
			},
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

	// Add object to lighting if it has brightness defined.
	if o.GetArchetype().Light != nil {
		gmap.AddObjectLighting(o, y, x, z)
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
		// Remove from lighting.
		if o.GetArchetype().Light != nil {
			gmap.RemoveObjectLighting(o, tile.Y, tile.X, tile.Z)
		}
		// Remove object.
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
	if o.GetArchetype().Light != nil {
		gmap.RemoveObjectLighting(o, o.GetTile().Y, o.GetTile().X, o.GetTile().Z)
	}

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

	// Update lighting.
	if o.GetArchetype().Light != nil {
		gmap.AddObjectLighting(o, o.GetTile().Y, o.GetTile().X, o.GetTile().Z)
	}

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

	// Remove old lighting.
	if o.GetArchetype().Light != nil {
		gmap.RemoveObjectLighting(o, o.GetTile().Y, o.GetTile().X, o.GetTile().Z)
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

	// Add new lighting.
	if o.GetArchetype().Light != nil {
		gmap.AddObjectLighting(o, o.GetTile().Y, o.GetTile().X, o.GetTile().Z)
	}

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

// QueueAction queues up an action for processing at the end of the current map update.
func (gmap *Map) QueueAction(action ActionI) {
	gmap.actions = append(gmap.actions, action)
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

// ShootCube shoots rays out in a cube centered on a location. onTileTouch is used to cancel the ray if it returns false. onRayEnd is called when a ray ceases.
func (gmap *Map) ShootCube(y1, x1, z1 float64, h, w, d float64, onTileTouch func(tile *Tile) bool, onRayEnd func()) {
	var c [][3]float64
	hh := h / 2
	wh := w / 2
	dh := d / 2
	ymin := y1 - hh
	if ymin < 0 {
		ymin = 0
	}
	ymax := y1 + hh
	if ymax > float64(gmap.height) {
		ymax = float64(gmap.height) - 1
	}

	xmin := x1 - wh
	if xmin < 0 {
		xmin = 0
	}
	xmax := x1 + wh
	if xmax > float64(gmap.width) {
		xmax = float64(gmap.width) - 1
	}

	zmin := z1 - dh
	if zmin < 0 {
		zmin = 0
	}
	zmax := z1 + dh
	if zmax > float64(gmap.depth) {
		zmax = float64(gmap.depth) - 1
	}

	for y := ymin; y < ymax; y += 2 {
		for x := xmin; x < xmax; x++ {
			for z := zmin; z < zmax; z++ {
				c = append(c, [3]float64{y, x, z})
			}
		}
	}

	for _, c := range c {
		end := false
		gmap.ShootRay(y1, x1, z1, c[0], c[1], c[2], func(tile *Tile) bool {
			r := onTileTouch(tile)
			if !r {
				onRayEnd()
				end = true
			}
			return r
		})
		if !end {
			onRayEnd()
		}
	}
}

// ShootRay shoots a ray from a position to another, calling f for each tile traversed. If f returns false, then the ray is stopped.
func (gmap *Map) ShootRay(fromY, fromX, fromZ, toY, toX, toZ float64, f func(tile *Tile) bool) (tiles []*Tile) {
	y1 := fromY
	x1 := fromX
	z1 := fromZ
	var tMaxX, tMaxY, tMaxZ, tDeltaX, tDeltaY, tDeltaZ float64
	y2 := toY
	x2 := toX
	z2 := toZ
	var dy, dx, dz int
	var y, x, z int

	sign := func(x float64) int {
		if x > 0 {
			return 1
		} else if x < 0 {
			return -1
		}
		return 0
	}
	frac0 := func(x float64) float64 {
		return x - math.Floor(x)
	}
	frac1 := func(x float64) float64 {
		return 1 - x + math.Floor(x)
	}

	dy = sign(y2 - y1)
	if dy != 0 {
		tDeltaY = math.Min(float64(dy)/(y2-y1), 10000)
	} else {
		tDeltaY = 10000
	}
	if dy > 0 {
		tMaxY = tDeltaY * frac1(y1)
	} else {
		tMaxY = tDeltaY * frac0(y1)
	}
	y = int(y1)

	dx = sign(x2 - x1)
	if dx != 0 {
		tDeltaX = math.Min(float64(dx)/(x2-x1), 10000)
	} else {
		tDeltaX = 10000
	}
	if dx > 0 {
		tMaxX = tDeltaX * frac1(x1)
	} else {
		tMaxX = tDeltaX * frac0(x1)
	}
	x = int(x1)

	dz = sign(z2 - z1)
	if dz != 0 {
		tDeltaZ = math.Min(float64(dz)/(z2-z1), 10000)
	} else {
		tDeltaZ = 10000
	}
	if dz > 0 {
		tMaxZ = tDeltaZ * frac1(z1)
	} else {
		tMaxZ = tDeltaZ * frac0(z1)
	}
	z = int(z1)

	for {
		if tMaxX < tMaxY {
			if tMaxX < tMaxZ {
				x += dx
				tMaxX += tDeltaX
			} else {
				z += dz
				tMaxZ += tDeltaZ
			}
		} else {
			if tMaxY < tMaxZ {
				y += dy
				tMaxY += tDeltaY
			} else {
				z += dz
				tMaxZ += tDeltaZ
			}
		}
		if tMaxY > 1 && tMaxX > 1 && tMaxZ > 1 {
			break
		}
		if y < 0 || x < 0 || z < 0 || y >= gmap.height || x >= gmap.width || z >= gmap.depth {
			continue
		}
		tile := gmap.GetTile(y, x, z)
		tiles = append(tiles, tile)
		if !f(tile) {
			break
		}
	}
	return
}

func (gmap *Map) RemoveObjectLighting(object ObjectI, y, x, z int) {
	if _, ok := gmap.lightObjects[object.GetID()]; ok {
		a := object.GetArchetype()
		h, w, d := object.GetDimensions()
		r := a.Light.Brightness / a.Light.Intensity
		v := a.Light.Brightness
		gmap.ShootCube(float64(y+h/2), float64(x+w/2), float64(z+d/2), float64(a.Light.Intensity), float64(a.Light.Intensity), float64(a.Light.Intensity), func(t *Tile) bool {
			t.removeObjectLight(object, v)
			v -= r
			if t.opaque || v == 0 {
				return false
			}
			return true
		}, func() {
			v = a.Light.Brightness
		})
		delete(gmap.lightObjects, object.GetID())
	} else {
		fmt.Println("FIXME: Removed lighting more than once")
	}
}

func (gmap *Map) AddObjectLighting(object ObjectI, y, x, z int) {
	if _, ok := gmap.lightObjects[object.GetID()]; !ok {
		a := object.GetArchetype()
		h, w, d := object.GetDimensions()
		r := a.Light.Brightness / a.Light.Intensity
		v := a.Light.Brightness
		gmap.ShootCube(float64(y+h/2), float64(x+w/2), float64(z+d/2), float64(a.Light.Intensity), float64(a.Light.Intensity), float64(a.Light.Intensity), func(t *Tile) bool {
			t.addObjectLight(object, v)
			v -= r
			if t.opaque || v == 0 {
				return false
			}
			return true
		}, func() {
			v = a.Light.Brightness
		})
		gmap.lightObjects[object.GetID()] = object
	} else {
		fmt.Println("FIXME: Added lighting more than once")
	}
}

// RefreshSky refreshes all tiles' sky field to match their exposure to the open sky.
func (gmap *Map) RefreshSky() {
	traversedCoords := make(map[[3]int]struct{}, 0)
	nextCoords := make([][3]int, 0)

	// First analyze/set all tiles that are guaranteed to be exposed to the sky.
	for x := 0; x < gmap.width; x++ {
		for z := 0; z < gmap.depth; z++ {
			for y := gmap.height - 1; y >= 0; y-- {
				// Bail if the tile is opaque.
				if gmap.tiles[y][x][z].opaque {
					break
				}

				gmap.tiles[y][x][z].sky = 1.0
				traversedCoords[[3]int{y, x, z}] = struct{}{}

				// Also add all adjacent tiles to our nextCoords slice that we will use to calculate their sky value.
				for y2 := -1; y2 < 2; y2 += 2 {
					for x2 := -1; x2 < 2; x2 += 2 {
						for z2 := -1; z2 < 2; z2 += 2 {
							if t := gmap.GetTile(y+y2, x+x2, z+z2); t != nil {
								// Don't re-analyze full-sky styles.
								if _, ok := traversedCoords[[3]int{y + y2, x + x2, z + z2}]; !ok {
									nextCoords = append(nextCoords, [3]int{y + y2, x + x2, z + z2})
								}
							}
						}
					}
				}

			}
		}
	}

	// This function iterates through each passed coordinate and makes it an aggregate of all adjacent tiles that have already been processed.
	processNextCoords := func(current [][3]int) [][3]int {
		nextCoords := make([][3]int, 0)
		for _, c := range current {
			t := gmap.GetTile(c[0], c[1], c[2])
			traversedCoords[c] = struct{}{}
			total := t.sky
			count := float32(1)

			// Check non-diagonals.
			targetCoords := [][3]int{
				{c[0] + 1, c[1], c[2]},
				{c[0] - 1, c[1], c[2]},
				{c[0], c[1], c[2] + 1},
				{c[0], c[1], c[2] - 1},
				{c[0], c[1] + 1, c[2]},
				{c[0], c[1] - 1, c[2]},
			}

			for _, target := range targetCoords {
				if t2 := gmap.GetTile(target[0], target[1], target[2]); t2 != nil {
					if _, ok := traversedCoords[target]; ok {
						if !t2.opaque {
							total += t2.sky
							count++
						}
					} else {
						// Add non-traversed coordinates to our next coordinates slice.
						nextCoords = append(nextCoords, target)
					}

				}
			}

			// Only adjust the sky value if it is less than 1.
			if t.sky < 1 {
				t.sky = total / count
				// Might as well round up numbers close enough to 1.
				if t.sky >= 0.8 {
					t.sky = 1
				}
			}
		}
		return nextCoords
	}

	// Step through all tile coordinates starting from full sky tiles.
	for c := processNextCoords(nextCoords); len(c) > 0; c = processNextCoords(c) {
	}
	// Dump the sky.
	/*for y := range gmap.tiles {
		for x := range gmap.tiles[y] {
			for z := range gmap.tiles[y][x] {
				if gmap.tiles[y][x][z].sky < 1.0 {
					fmt.Printf("Sky %dx%dx%d: %f\n", y, x, z, gmap.tiles[y][x][z].sky)
				}
			}
		}
	}*/
}
