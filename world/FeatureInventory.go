package world

import (
	"errors"
)

// FeatureInventory provides the ability to store objects. It is used for the player's basic inventory system as well as any storage, such as bags, beltpouches, etc., that the character may have. The capacity and space are defined on owning object creation. In general capacity/maxCapacity is only ever updated when an owning character increase their physical attributes. Maximum volume should remain the same from the object's creation, with the possible exception of if the owning object is a character that grows in size.
type FeatureInventory struct {
	inventory   []ObjectI
	changed     bool
	maxCapacity float64 // The max capacity that can be carried.
	capacity    float64 // The current capacity that is available.
	maxVolume   int     // Maximum volume that can be stored. This is maxWidth*maxHeight*maxDepth
	volume      int     // The current volume available to objects.
}

// Errors
var (
	ErrObjectMissingInInventory = errors.New("object does not exist in inventory")
	ErrObjectAlreadyInInventory = errors.New("object already exists in inventory")
	ErrObjectTooHeavy           = errors.New("object is too heavy")
	ErrObjectTooLarge           = errors.New("object is too large")
)

// SetDimensions sets the maximum and current container volume.
func (f *FeatureInventory) SetDimensions(h, w, d int) {
	f.maxVolume = h * w * d
	f.volume = f.maxVolume
}

// AddInventoryObject directly adds the given object to the inventory.
func (f *FeatureInventory) AddInventoryObject(o ObjectI) error {
	for _, v := range f.inventory {
		if v == o {
			return ErrObjectAlreadyInInventory
		}
	}
	h, w, d := o.GetDimensions()
	v := h * w * d
	if f.volume-v < 0 {
		return ErrObjectTooLarge
	}
	// TODO: Check weight of object.

	f.volume -= v
	f.inventory = append(f.inventory, o)
	f.changed = true
	return nil
}

// RemoveInventoryObject removes the given object from the inventory.
func (f *FeatureInventory) RemoveInventoryObject(o ObjectI) error {
	for i, v := range f.inventory {
		if v == o {
			h, w, d := o.GetDimensions()
			f.volume += h * w * d
			f.inventory = append(f.inventory[:i], f.inventory[i+1:]...)
			f.changed = true
			return nil
		}
	}
	return ErrObjectMissingInInventory
}

// FindInventory finds the given inventory that matches the cb.
func (f *FeatureInventory) FindInventory(cb func(v ObjectI) bool) (matches []ObjectI) {
	for _, o := range f.inventory {
		if cb(o) {
			matches = append(matches, o)
		}
	}
	return
}
