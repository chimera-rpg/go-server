package world

import (
	"errors"
	"time"
)

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

func (m *Map) HandleActionEquip(a *ActionEquip) error {
	if o, ok := a.object.(*ObjectCharacter); ok {
		// FIXME: Allow non ObjectCharacters to be containers.
		if targetContainer := m.world.GetObject(a.Container); targetContainer != nil {
			if targetContainer, ok := targetContainer.(FeatureInventoryI); ok {
				targetObject, err := targetContainer.GetObjectByID(a.Target)
				if err != nil {
					return err
				}
				// Attempt to equip it.
				if err := o.Equip(targetObject); err != nil {
					return err
				}
				// Remove it from the inventory.
				if err := o.RemoveInventoryObject(targetObject); err != nil {
					return err
				}
				o.SetContainer(nil)
			}
		}
		// TODO: Send message that it was equipped!
	}
	return errors.New("object is not a character")
}
