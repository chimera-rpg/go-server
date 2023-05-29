package world

import (
	"errors"

	"github.com/chimera-rpg/go-server/data"
)

// FeatureEquipment provides the ability to equip equipable objects.
type FeatureEquipment struct {
	equipment []ObjectI
	// slots will be a pointer to the owning object or archetype's slots field.
	slots   *data.Slots
	changed bool
}

// Errors
var (
	ErrObjectEquipmentMissing = errors.New("object does not exist in equipment")
	ErrObjectCannotEquip      = errors.New("cannot equip object")
	ErrObjectNotEquipped      = errors.New("object is not equipped")
)

// AddEquipmentObject directly adds the given object to the equipment.
func (f *FeatureEquipment) AddEquipmentObject(o ObjectI) bool {
	if o, ok := o.(*ObjectEquipable); ok {
		f.equipment = append(f.equipment, o)
		return true
	}
	return false
}

// CanEquip returns if the object can be equipped. FIXME: Make this return an error so we can provide a message to the user saying why they couldn't equip the item.
func (f *FeatureEquipment) CanEquip(ob ObjectI) bool {
	if _, ok := ob.(*ObjectEquipable); !ok {
		return false
	}
	// Check the object's uses against our free slots.
	for k, v := range ob.GetArchetype().Slots.UsesIDs {
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
	for k, v := range ob.GetArchetype().Slots.Needs.MinIDs {
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
	for k, v := range ob.GetArchetype().Slots.Needs.MaxIDs {
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
func (f *FeatureEquipment) Equip(ob ObjectI) error {
	if !f.CanEquip(ob) {
		return ErrObjectCannotEquip
	}

	for k, v := range ob.GetArchetype().Slots.Uses {
		f.slots.Free[k] -= v
	}

	if f.AddEquipmentObject(ob) {
		f.equipment = append(f.equipment, ob)
	}

	f.changed = true

	return nil
}

// Unequip attempts to remove the given object from the equipment slice.
func (f *FeatureEquipment) Unequip(ob ObjectI) error {
	for i, v := range f.equipment {
		if v == ob {
			f.equipment = append(f.equipment[:i], f.equipment[i+1:]...)

			for k, v := range ob.GetArchetype().Slots.Uses {
				f.slots.Free[k] += v
			}

			f.changed = true
			return nil
		}
	}
	return ErrObjectNotEquipped
}

// FindEquipment finds the given equipment that matches the cb.
func (f *FeatureEquipment) FindEquipment(cb func(v ObjectI) bool) (matches []ObjectI) {
	for _, o := range f.equipment {
		if cb(o) {
			matches = append(matches, o)
		}
	}
	return
}

func (f *FeatureEquipment) GetObjectByID(id ID) (ObjectI, error) {
	for _, o := range f.equipment {
		if o.GetID() == id {
			return o, nil
		}
	}
	return nil, ErrObjectMissingInInventory
}
