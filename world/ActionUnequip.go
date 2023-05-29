package world

import "time"

// ActionUnequip represents an action to unequip a given object.
type ActionUnequip struct {
	Action
	Target    ID
	Container ID
}

// NewActionUnequip returns an instantized version of ActionUnequip.
func NewActionUnequip(container ID, target ID, cost time.Duration) *ActionEquip {
	return &ActionEquip{
		// FIXME: Figure out the actual base costs of unequipping. Probably based on the weight + size of the item vs. the character's own.
		Action: Action{
			channel:  cost / 4,
			recovery: cost - cost/4,
		},
		Target:    target,
		Container: container,
	}
}
