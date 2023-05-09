package world

import (
	"errors"

	"github.com/chimera-rpg/go-server/data"
)

type FeatureInventory struct {
	inventory        []ObjectI
	equipment        []*ObjectEquipable
	slots            *data.Slots
	inventoryChanged bool
	equipmentChanged bool
}

// Errors
var (
	ErrObjectMissingInInventory = errors.New("object does not exist in inventory")
	ErrObjectCannotEquip        = errors.New("cannot equip object")
	ErrObjectNotEquipped        = errors.New("object is not equipped")
)

// AddInventoryObject directly adds the given object to the inventory.
func (f *FeatureInventory) AddInventoryObject(o ObjectI) bool {
	f.inventory = append(f.inventory, o)
	return true
}

// AddEquipmentObject directly adds the given object to the equipment.
func (f *FeatureInventory) AddEquipmentObject(o ObjectI) bool {
	if o, ok := o.(*ObjectEquipable); ok {
		f.equipment = append(f.equipment, o)
		return true
	}
	return false
}

// CanEquip returns if the object can be equipped. FIXME: Make this return an error so we can provide a message to the user saying why they couldn't equip the item.
func (f *FeatureInventory) CanEquip(ob *ObjectEquipable) bool {
	// Check the object's uses against our free slots.
	for k, v := range ob.Archetype.Slots.UsesIDs {
		v2, ok := f.slots.FreeIDs[k]
		if !ok {
			// No such slot is available.
			return false
		}
		if v2 < v {
			// We have the slot, but are missing v - v2 count.
			return false
		}
	}

	// Check for minimum slot requirements.
	for k, v := range ob.Archetype.Slots.Needs.MinIDs {
		v2, ok := f.slots.HasIDs[k]
		if !ok {
			// No such slot exists.
			return false
		}
		if v2 < v {
			// We have the slot, but are missing v - v2 count.
			return false
		}
	}
	// Check for maximum slot requirements.
	for k, v := range ob.Archetype.Slots.Needs.MaxIDs {
		v2, ok := f.slots.HasIDs[k]
		if !ok {
			// No such slot exists.
			return false
		}
		if v2 > v {
			// We have the slot, but are in excess by v2 - v
			return false
		}
	}

	return true
}

// Equip attempts to equip the given object.
func (f *FeatureInventory) Equip(ob *ObjectEquipable) error {
	index := -1
	for i, v := range f.inventory {
		if v == ob {
			index = i
			break
		}
	}
	if index == -1 {
		return ErrObjectMissingInInventory
	}

	if !f.CanEquip(ob) {
		return ErrObjectCannotEquip
	}

	for k, v := range ob.Archetype.Slots.Uses {
		f.slots.Free[k] -= v
	}

	if f.AddEquipmentObject(f.inventory[index]) {
		f.inventory = append(f.inventory[:index], f.inventory[index+1:]...)
	}

	f.equipmentChanged = true

	return nil
}

// Unequip attempts to remove the given object from the equipment slice and into the inventory slice.
func (f *FeatureInventory) Unequip(ob *ObjectEquipable) error {
	for i, v := range f.equipment {
		if v == ob {
			f.equipment = append(f.equipment[:i], f.equipment[i+1:]...)
			f.AddInventoryObject(v)

			for k, v := range ob.Archetype.Slots.Uses {
				f.slots.Free[k] += v
			}

			f.equipmentChanged = true
			return nil
		}
	}
	return ErrObjectNotEquipped
}

// FindEquipment finds the given equipment that matches the cb.
func (f *FeatureInventory) FindEquipment(cb func(v ObjectI) bool) (matches []*ObjectEquipable) {
	for _, o := range f.equipment {
		if cb(o) {
			matches = append(matches, o)
		}
	}
	return
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
