package world

import (
	"errors"
	"time"
)

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

func (m *Map) HandleActionDrop(a *ActionDrop) error {
	if o, ok := a.object.(*ObjectCharacter); ok {
		fromContainer, err := o.getContainerOf(a.FromContainer, a.Target)
		if err != nil {
			return err
		}
		if fromContainer == nil {
			// We're just moving a given world item to a new location.
			target := o.tile.gameMap.world.GetObject(a.Target)
			if target == nil {
				return errors.New("target to drop does not exist")
			}
			blocks, err := target.GetTile().gameMap.IsPositionOpen(target, a.Y, a.X, a.Z)
			if err != nil {
				// TODO: Send message to player to let them know the target position is invalid.
				return err
			}
			if blocks {
				// TODO: Send message to player to let them know the target position is invalid.
				return errors.New("target position blocked")
			}
			if err := target.GetTile().gameMap.TeleportObject(target, a.Y, a.X, a.Z, false); err != nil {
				// TODO: Send message to player to let them know moving failed.
				return err
			}
		} else {
			// We're dropping from a container.
			target, err := fromContainer.(FeatureInventoryI).GetObjectByID(a.Target)
			if err != nil {
				return err
			}

			blocks, err := o.GetTile().gameMap.IsPositionOpen(target, a.Y, a.X, a.Z)
			if err != nil {
				// TODO: Send message to player to let them know the target position is invalid.
				return err
			}
			if blocks {
				// TODO: Send message to player to let them know the target position is invalid.
				return errors.New("target position blocked")
			}
			// Place object.
			if err := o.GetTile().gameMap.PlaceObject(target, a.Y, a.X, a.Z); err != nil {
				// TODO: Send message to player to let them know the target failed to be placed.
				return err
			}
			target.SetContainer(nil)
			// Remove from container.
			if err := fromContainer.(FeatureInventoryI).RemoveInventoryObject(target); err != nil {
				// TODO: Send message that the item could not be removed from inventory.
				return err
			}
		}
		// TODO: notify of move?
	}
	return errors.New("object is not a character")
}
