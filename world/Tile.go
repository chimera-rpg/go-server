package world

import (
	"errors"
)

// Tile represents a location on the ground.
type Tile struct {
	gameMap     *Map      // I guess this okay.
	y, x, z     int       // Location of the tile.
	objects     []ObjectI // objects contains Objects that origin from this tile. This data is used in network transmission.
	objectParts []ObjectI // objectParts contains Object pointers that are used for collisions, pathing, and otherwise. This data is never sent over the network.
	brightness  int
	modTime     uint16 // Last time this tile was updated.
}

// insertObject inserts the provided Object at the given index.
func (tile *Tile) insertObject(object ObjectI, index int) error {
	if index == -1 {
		index = len(tile.objects)
	}
	if index == -1 {
		index = 0
	}
	if object.GetTile() != nil {
		object.GetTile().removeObject(object)
	}

	tile.objects = append(tile.objects[:index], append([]ObjectI{object}, tile.objects[index:]...)...)

	// Update object's tile reference.
	object.SetTile(tile)

	tile.modTime++

	return nil
}

func (tile *Tile) removeObject(object ObjectI) error {
	for i := range tile.objects {
		if tile.objects[i] == object {
			tile.objects = append(tile.objects[:i], tile.objects[i+1:]...)
			object.SetTile(nil)
			tile.modTime++
			return nil
		}
	}
	return errors.New("object to remove does not exist")
}

// GetObjects returns a slice of the tile's Object interfaces.
func (tile *Tile) GetObjects() []ObjectI {
	return tile.objects
}

// GetMap returns the owning map of the Tile.
func (tile *Tile) GetMap() *Map {
	return tile.gameMap
}
