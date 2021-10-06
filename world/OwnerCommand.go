package world

// OwnerCommand does stuff.
type OwnerCommand interface {
}

// OwnerMoveCommand represents an owner command to move an object towards a given direction.
type OwnerMoveCommand struct {
	Y, X, Z int
	Running bool // Whether or not it is a running move.
}

// OwnerStatusCommand represents an owner command to set the status for its target object.
type OwnerStatusCommand struct {
	Status StatusI
}
