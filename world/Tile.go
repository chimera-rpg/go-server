package world

import (
	"errors"
)

// Tile represents a location on the ground.
type Tile struct {
	y, x, z    int // Location of the tile.
	objects    []ObjectI
	brightness int
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

	return nil
}

func (tile *Tile) removeObject(object ObjectI) error {
	for i := range tile.objects {
		if tile.objects[i] == object {
			tile.objects = append(tile.objects[:i], tile.objects[i+1:]...)
			object.SetTile(nil)
			return nil
		}
	}
	return errors.New("object to remove does not exist")
}

func (tile *Tile) GetObjects() []ObjectI {
	return tile.objects
}
