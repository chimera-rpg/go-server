package world

import (
	"errors"
	"time"
)

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

func (m *Map) HandleActionGrab(a *ActionGrab) error {
	if o, ok := a.object.(*ObjectCharacter); ok {
		fromContainer, err := o.getContainerOf(a.Target, a.FromContainer)
		if err != nil {
			return err
		}
		toContainer, err := o.getContainer(a.ToContainer)
		if err != nil {
			return err
		}
		var targetObject ObjectI

		// 1. Get the target object from either the world or its container.
		if fromContainer == nil {
			// If from container is nil, then it is in the world.
			targetObject = o.tile.gameMap.world.GetObject(a.Target)
			if targetObject == nil {
				return errors.New("object not in world")
			}
		} else {
			// Otherwise it's in the given container.
			targetObject, err = fromContainer.(FeatureInventoryI).GetObjectByID(a.Target)
			if err != nil {
				return err
			}
		}
		// 2. Add the object to the given container.
		if err := toContainer.(FeatureInventoryI).AddInventoryObject(targetObject); err != nil {
			if err == ErrObjectAlreadyInInventory {
				// Somehow already in the player inventory...
				return err
			} else if err == ErrObjectTooLarge {
				// TODO: Send information to player that it won't fit!
				return err
			}
		} else {
			targetObject.SetContainer(toContainer)
			// TODO: Send information to the player that it was added!
		}
		// 3. Remove the object from the world or its container.
		if fromContainer == nil {
			if err := targetObject.GetTile().GetMap().RemoveObject(targetObject); err != nil {
				return err
			}
		} else {
			if err := fromContainer.(FeatureInventoryI).RemoveInventoryObject(targetObject); err != nil {
				return err
			}
		}
	}
	return errors.New("object is not a character")
}
