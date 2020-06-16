package world

// Tile represents a location on the ground.
type Tile struct {
	objects    []ObjectI
	brightness int
}

// insertObject inserts the provided Object at the given index.
func (tile *Tile) insertObject(object ObjectI, index int) error {
	if len(tile.objects) == index {
		tile.objects = append(tile.objects, object)
	}
	tile.objects = append(tile.objects[:index+1], tile.objects[index:]...)
	tile.objects[index] = object

	return nil
}
