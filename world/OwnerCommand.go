package world

// OwnerCommand does stuff.
type OwnerCommand interface {
}

// OwnerMoveCommand represents an owner command to move an object towards a given direction.
type OwnerMoveCommand struct {
	Y, X, Z int
}
