package world

// OwnerCommand does stuff.
type OwnerCommand interface {
}

type OwnerMoveCommand struct {
	Y, X, Z int
}
