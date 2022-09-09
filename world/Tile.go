package world

import (
	"errors"

	cdata "github.com/chimera-rpg/go-common/data"
)

type TileLight struct {
	o       ObjectI
	R, G, B uint8
}

// Tile represents a location on the ground.
type Tile struct {
	gameMap      *Map      // I guess this okay.
	Y, X, Z      int       // Location of the tile.
	objects      []ObjectI // objects contains Objects that origin from this tile. This data is used in network transmission.
	objectParts  []ObjectI // objectParts contains Object pointers that are used for collisions, pathing, and otherwise. This data is never sent over the network.
	lights       []TileLight
	r, g, b      uint8 // r, g, b are the final computed values from lights.
	blocking     cdata.MatterType
	matter       cdata.MatterType
	opaque       bool
	modTime      uint16  // Last time this tile was updated.
	lightModTime uint16  // Last time this tile's light was updated.
	skyModTime   uint16  // Last time this tile's sky was updated.
	sky          float32 // How much this time is considered to be exposed to the open sky. Only calculated on map creation (for now).
	haven        bool    // Whether this tile is considered as a safe location for the player to disconnect and their character to be saved in.
}

// insertObject inserts the provided Object at the given index.
func (tile *Tile) insertObject(object ObjectI, index int) error {
	// Remove from old tile _first_ just in case the object is in this tile already.
	if object.GetTile() != nil {
		object.GetTile().removeObject(object)
	}

	if index == -1 {
		index = len(tile.objects)
	}
	if index == -1 {
		index = 0
	}

	if len(tile.objects) == 0 {
		tile.objects = append(tile.objects, object)
	} else {
		tile.objects = append(tile.objects[:index], append([]ObjectI{object}, tile.objects[index:]...)...)
	}

	// Update object's tile reference.
	object.SetTile(tile)

	tile.updateStates()
	tile.modTime++

	return nil
}

func (tile *Tile) removeObject(object ObjectI) error {
	i := tile.getObjectIndex(object)
	if i >= 0 {
		tile.objects = append(tile.objects[:i], tile.objects[i+1:]...)
		object.SetTile(nil)
		tile.modTime++
		tile.updateStates()
		return nil
	}
	return errors.New("object to remove does not exist")
}

func (tile *Tile) insertObjectPart(object ObjectI, index int) {
	if index == -1 {
		index = len(tile.objectParts)
	}
	if index == -1 {
		index = 0
	}

	if existingIndex := tile.getObjectPartIndex(object); existingIndex == -1 {
		if len(tile.objectParts) == 0 {
			tile.objectParts = append(tile.objectParts, object)
		} else {
			tile.objectParts = append(tile.objectParts[:index], append([]ObjectI{object}, tile.objectParts[index:]...)...)
		}
	}

	tile.updateStates()
}

// removeObjectPart removes a collision object reference.
func (tile *Tile) removeObjectPart(object ObjectI) {
	i := tile.getObjectPartIndex(object)
	if i >= 0 {
		tile.objectParts = append(tile.objectParts[:i], tile.objectParts[i+1:]...)
	}
	tile.updateStates()
}

func (tile *Tile) getObjectPartIndex(object ObjectI) int {
	for i, o := range tile.objectParts {
		if o.GetID() == object.GetID() {
			return i
		}
	}
	return -1
}

func (tile *Tile) getObjectIndex(object ObjectI) int {
	for i, o := range tile.objects {
		if o.GetID() == object.GetID() {
			return i
		}
	}
	return -1
}

func (tile *Tile) getObjectLightIndex(object ObjectI) int {
	for i, l := range tile.lights {
		if l.o.GetID() == object.GetID() {
			return i
		}
	}
	return -1
}

// GetObjects returns a slice of the tile's Object interfaces.
func (tile *Tile) GetObjects() []ObjectI {
	return tile.objects
}

// GetMap returns the owning map of the Tile.
func (tile *Tile) GetMap() *Map {
	return tile.gameMap
}

// CheckObjects calls the given function on all contained objects and returns true as soon as the function returns true.
func (tile *Tile) CheckObjects(f func(ObjectI) bool) bool {
	for _, o := range tile.objects {
		if f(o) {
			return true
		}
	}
	return false
}

// updateStates updates various cached states of the tile, such as blocking or haven.
func (tile *Tile) updateStates() {
	tile.matter = 0
	tile.blocking = 0
	tile.opaque = false
	tile.haven = false
	for _, o := range tile.objects {
		a := o.GetArchetype()
		tile.blocking |= a.Blocking
		tile.matter |= o.Matter()
		if a.Matter.Is(cdata.OpaqueMatter) {
			tile.opaque = true
		}
		if a.Specials.Haven {
			tile.haven = true
		}
	}
}

func getUniqueObjectsInTiles(tiles []*Tile) (objs []ObjectI) {
	for _, t := range tiles {
		for _, o := range t.objectParts {
			exists := false
			for _, o2 := range objs {
				if o == o2 {
					exists = true
					break
				}
			}
			if !exists {
				objs = append(objs, o)
			}
		}
	}
	return
}

func (tile *Tile) addObjectLight(object ObjectI, r, g, b uint8) {
	i := tile.getObjectLightIndex(object)
	if i == -1 {
		tile.lights = append(tile.lights, TileLight{
			o: object,
			R: r,
			G: g,
			B: b,
		})
		tile.calculateLight()
	}
}

func (tile *Tile) removeObjectLight(object ObjectI, r, g, b uint8) {
	i := tile.getObjectLightIndex(object)
	if i >= 0 {
		tile.lights = append(tile.lights[:i], tile.lights[i+1:]...)
		tile.calculateLight()
	}
}

func (tile *Tile) calculateLight() {
	tile.lightModTime++
	r := uint32(0)
	g := uint32(0)
	b := uint32(0)
	if len(tile.lights) > 0 {
		var maxR, maxG, maxB uint8
		for _, l := range tile.lights {
			if l.R > maxR {
				maxR = l.R
			}
			if l.G > maxG {
				maxG = l.G
			}
			if l.B > maxB {
				maxB = l.B
			}

			r += uint32(l.R)
			g += uint32(l.G)
			b += uint32(l.B)
		}
		r = (r-uint32(maxR))/uint32(len(tile.lights)) + uint32(maxR)
		g = (g-uint32(maxG))/uint32(len(tile.lights)) + uint32(maxG)
		b = (b-uint32(maxB))/uint32(len(tile.lights)) + uint32(maxB)
	}
	if r >= 255 {
		tile.r = 255
	} else {
		tile.r = uint8(r)
	}
	if g >= 255 {
		tile.g = 255
	} else {
		tile.g = uint8(g)
	}
	if b >= 255 {
		tile.b = 255
	} else {
		tile.b = uint8(b)
	}
}
