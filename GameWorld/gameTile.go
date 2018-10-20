package GameWorld

type gameTile struct {
  object *GameObject
  brightness int
}

func (tile *gameTile) insertObject(object *GameObject, index int) error {
  if tile.object == nil {
    tile.object = object
    return nil
  }
  target := tile.object

  if index == 0 {
    target.previous = object
    object.next = target
    tile.object = object
    return nil
  } else if index > 0 {
    for i := 0; target.next != nil && i != index; i, target = i+1, target.next {}
  } else if index < 0 {
    // -2, so (i = count, i != count-index)
    // Set target to the end
    count := 0
    for ; target.next != nil; target = target.next { count++ }
    if count-index < 0 {
      index = 0
    }
    // Now iterate backwards until we find the appropriate position
    for i := count; target.previous != nil && i != count-index; i, target = i-1, target.previous {}
    //
  }
  if target.next != nil {
    object.next = target.next
  }
  target.next = object
  object.removeSelf()
  object.previous = target
  return nil
}
