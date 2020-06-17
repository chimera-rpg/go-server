package world

import (
	"errors"
)

// Tile represents a location on the ground.
type Tile struct {
	objects    []ObjectI
	brightness int
}

// insertObject inserts the provided Object at the given index.
func (tile *Tile) insertObject(object ObjectI, index int) error {
	if object.GetTile() != nil {
		object.GetTile().removeObject(object)
	}
	if len(tile.objects) == index {
		tile.objects = append(tile.objects, object)
	}
	tile.objects = append(tile.objects[:index+1], tile.objects[index:]...)
	tile.objects[index] = object

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
