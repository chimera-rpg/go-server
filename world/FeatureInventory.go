package world

import (
	"errors"
)

// FeatureInventory provides the ability to store objects. It is used for the player's basic inventory system as well as any storage, such as bags, beltpouches, etc., that the character may have. The capacity and space are defined on owning object creation. In general capacity/maxCapacity is only ever updated when an owning character increase their physical attributes. Maximum volume should remain the same from the object's creation, with the possible exception of if the owning object is a character that grows in size.
type FeatureInventory struct {
	inventory   []ObjectI
	changed     bool
	maxCapacity float64 // The max capacity in kg that can be carried.
	capacity    float64 // The current capacity in kg that is available.
	maxVolume   int     // Maximum volume that can be stored. This is maxWidth*maxHeight*maxDepth
	volume      int     // The current volume available to objects.
}

// Errors
var (
	ErrObjectMissingInInventory = errors.New("object does not exist in inventory")
	ErrObjectAlreadyInInventory = errors.New("object already exists in inventory")
	ErrObjectTooHeavy           = errors.New("object is too heavy")
	ErrObjectTooLarge           = errors.New("object is too large")
	ErrObjectOverVolume         = errors.New("inventory is over volume")
	ErrObjectOverCapacity       = errors.New("inventory is over capacity")
)

// SetVolume sets the maximum and current container volume.
func (f *FeatureInventory) SetVolume(v int) (err error) {
	if v > f.maxVolume {
		f.volume += v - f.maxVolume
	} else if v < f.maxVolume {
		err = ErrObjectOverVolume
		// TODO: Flag the inventory to dump enough of its contents to not be over volume.
	}
	f.maxVolume = v
	return
}

// SetCapacity sets the maximum capacity and adjusts the current.
func (f *FeatureInventory) SetCapacity(v float64) (err error) {
	if v > f.maxCapacity {
		f.capacity += v - f.maxCapacity
	} else if v < f.maxCapacity {
		// TODO: Flag the inventory to dump enough of its contents to not be over capacity.
		err = ErrObjectOverCapacity
	}
	f.maxCapacity = v
	return
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

func (f *FeatureInventory) GetObjectByID(id ID) (ObjectI, error) {
	for _, o := range f.inventory {
		if o.GetID() == id {
			return o, nil
		}
	}
	return nil, ErrObjectMissingInInventory
}

type FeatureInventoryI interface {
	GetObjectByID(ID) (ObjectI, error)
	AddInventoryObject(ObjectI) error
	RemoveInventoryObject(ObjectI) error
}
