package world

import "time"

// ActionGrab represents an action to grab a given object.
type ActionGrab struct {
	Action
	FromContainer ID // Container to grab from.
	ToContainer   ID // Container to place the item into.
	Target        ID // Target to grab.
}

// NewActionGrab returns an instantized version of ActionGrab.
func NewActionGrab(from, to ID, target ID, cost time.Duration) *ActionGrab {
	return &ActionGrab{
		// FIXME: Figure out the actual base costs of grabbing. Probably based on the weight + size of the item vs. the character's own strength.
		Action: Action{
			channel:  cost / 4,
			recovery: cost - cost/4,
		},
		FromContainer: from,
		ToContainer:   to,
		Target:        target,
	}
}
