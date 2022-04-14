package world

import (
	"errors"

	cdata "github.com/chimera-rpg/go-common/data"
)

// Tile represents a location on the ground.
type Tile struct {
	gameMap      *Map      // I guess this okay.
	Y, X, Z      int       // Location of the tile.
	objects      []ObjectI // objects contains Objects that origin from this tile. This data is used in network transmission.
	objectParts  []ObjectI // objectParts contains Object pointers that are used for collisions, pathing, and otherwise. This data is never sent over the network.
	objectLights []ObjectI // objectLights contains Objects that give light to this tile.
	brightness   float32
	r, g, b      uint8
	blocking     cdata.MatterType
	matter       cdata.MatterType
	opaque       bool
	modTime      uint16 // Last time this tile was updated.
	lightModTime uint16 // Last time this tile's light was updated.
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

	tile.updateBlocking()
	tile.modTime++

	return nil
}

func (tile *Tile) removeObject(object ObjectI) error {
	i := tile.getObjectIndex(object)
	if i >= 0 {
		tile.objects = append(tile.objects[:i], tile.objects[i+1:]...)
		object.SetTile(nil)
		tile.modTime++
		tile.updateBlocking()
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

	tile.updateBlocking()
}

// removeObjectPart removes a collision object reference.
func (tile *Tile) removeObjectPart(object ObjectI) {
	i := tile.getObjectPartIndex(object)
	if i >= 0 {
		tile.objectParts = append(tile.objectParts[:i], tile.objectParts[i+1:]...)
	}
	tile.updateBlocking()
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
	for i, o := range tile.objectLights {
		if o.GetID() == object.GetID() {
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

func (tile *Tile) updateBlocking() {
	tile.matter = 0
	tile.blocking = 0
	tile.opaque = false
	for _, o := range tile.objects {
		a := o.GetArchetype()
		tile.blocking |= a.Blocking
		tile.matter |= o.Matter()
		if a.Matter.Is(cdata.OpaqueMatter) {
			tile.opaque = true
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

func (tile *Tile) addObjectLight(object ObjectI, brightness float32) {
	i := tile.getObjectLightIndex(object)
	if i == -1 {
		tile.objectLights = append(tile.objectLights, object)
		tile.brightness += brightness
		tile.lightModTime++
	}
}

func (tile *Tile) removeObjectLight(object ObjectI, brightness float32) {
	i := tile.getObjectLightIndex(object)
	if i >= 0 {
		tile.objectLights = append(tile.objectLights[:i], tile.objectLights[i+1:]...)
		tile.brightness -= brightness
		tile.lightModTime++
	}
}
