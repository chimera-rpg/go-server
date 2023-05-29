package world

import "time"

// ActionDrop represents an action to drop a given object into the world.
type ActionDrop struct {
	Action
	FromContainer ID  // Container to grab from.
	Y, X, Z       int // Position to drop to.
	Target        ID  // Target to grab.
}

// NewActionDrop returns an instantized version of ActionDrop.
func NewActionDrop(from ID, y, x, z int, target ID, cost time.Duration) *ActionDrop {
	return &ActionDrop{
		// FIXME: Figure out the actual base costs of grabbing. Probably based on the weight + size of the item vs. the character's own strength.
		Action: Action{
			channel:  cost / 4,
			recovery: cost - cost/4,
		},
		FromContainer: from,
		Y:             y,
		X:             x,
		Z:             z,
		Target:        target,
	}
}
