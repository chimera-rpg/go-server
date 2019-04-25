package world

// Tile represents a location on the ground.
type Tile struct {
	object     ObjectI
	brightness int
}

// insertObject inserts the provided Object at the given index.
func (tile *Tile) insertObject(object ObjectI, index int) error {
	if tile.object == nil {
		tile.object = object
		return nil
	}
	target := tile.object

	if index == 0 {
		target.setPrevious(object)
		object.setNext(target)
		tile.object = object
		return nil
	} else if index > 0 {
		for i := 0; target.getNext() != nil && i != index; i, target = i+1, target.getNext() {
		}
	} else if index < 0 {
		// -2, so (i = count, i != count-index)
		// Set target to the end
		count := 0
		for ; target.getNext() != nil; target = target.getNext() {
			count++
		}
		if count-index < 0 {
			index = 0
		}
		// Now iterate backwards until we find the appropriate position
		for i := count; target.getPrevious() != nil && i != count-index; i, target = i-1, target.getPrevious() {
		}
		//
	}
	if target.getNext() != nil {
		object.setNext(target.getNext())
	}
	target.setNext(object)
	object.removeSelf()
	object.setPrevious(target)
	return nil
}
