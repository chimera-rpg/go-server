package world

import "time"

// ActionEquip represents an action to equip a given object.
type ActionEquip struct {
	Action
	Container          ID
	ContainerInventory *FeatureInventory
	Target             ID
}

// NewActionEquip returns an instantized version of ActionEquip.
func NewActionEquip(container ID, target ID, cost time.Duration) *ActionEquip {
	return &ActionEquip{
		// FIXME: Figure out the actual base costs of equipping. Probably based on the weight + size of the item vs. the character's own.
		Action: Action{
			channel:  cost / 4,
			recovery: cost - cost/4,
		},
		Container: container,
		Target:    target,
	}
}
