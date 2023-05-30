package world

import (
	"errors"
	"time"
)

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

func (m *Map) HandleActionUnequip(a *ActionUnequip) error {
	if o, ok := a.object.(*ObjectCharacter); ok {
		// The inventory to place the target in.
		var targetInventory ObjectI
		// Find the given container to add the equipment to.
		if a.Container != 0 {
			container := m.world.GetObject(a.Container)
			if container != nil {
				// TODO: Do a bounds check to ensure the given container is in reach of the player and isn't locked/otherwise.
				if _, ok := container.(FeatureInventoryI); ok {
					targetInventory = container
				}
			}
		}
		// Default to character's own inventory as the target container if one is not set.
		if targetInventory == nil {
			targetInventory = o
		}
		// Get the equipment to remove.
		targetObject, err := o.FeatureEquipment.GetObjectByID(a.Target)
		if err != nil {
			return err
		}
		// Remove the object from the equipment.
		if err := o.FeatureEquipment.Unequip(targetObject); err != nil {
			return err
		}
		// Attempt to add the object to the inventory.
		if err := targetInventory.(FeatureInventoryI).AddInventoryObject(targetObject); err != nil {
			// Drop item on the tile that the player is on if an error happens when adding to the given inventory.
			if err := m.PlaceObject(targetObject, o.tile.Y, o.tile.X, o.tile.Z); err != nil {
				return errors.Join(errors.New("item is lost"), err)
			}
			return err
		}
		o.SetContainer(targetInventory)
		// TODO: Send update that it was unequipped!
	}
	return errors.New("object is not a character")
}
