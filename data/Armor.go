package data

// AttackType represents the % damage type that a weapon does or armor protects from.
type AttackType float32

// Armors is our collection of damage types, used by both armor and weapons.
type AttackTypes struct {
	// Overarching damage types.
	Physical AttackType `yaml:"Physical,omitempty"`
	Arcane   AttackType `yaml:"Arcane,omitempty"`
	Spirit   AttackType `yaml:"Spirit,omitempty"`
	// Physical Subtypes.
	Impact AttackType `yaml:"Impact,omitempty"`
	Pierce AttackType `yaml:"Pierce,omitempty"`
	Edged  AttackType `yaml:"Edged,omitempty"`
	// Arcane Subtypes.
	Flame     AttackType `yaml:"Flame,omitempty"`
	Frost     AttackType `yaml:"Frost,omitempty"`
	Lightning AttackType `yaml:"Lightning,omitempty"`
	Corrosive AttackType `yaml:"Corrosive,omitempty"`
	Force     AttackType `yaml:"Force,omitempty"`
	// Spirit Subtypes.
	Heal AttackType `yaml:"Heal,omitempty"`
	Harm AttackType `yaml:"Harm,omitempty"`
	// TODO: Figure out our actual full list of damage types.
}

// WeaponType represents the % weapon type that the weapon does in physical damage.
type WeaponType float32

// WeaponTypes is our collection of weapon types.
type WeaponTypes struct {
	Impact WeaponType `yaml:"Impact,omitempty"`
	Pierce WeaponType `yaml:"Pierce,omitempty"`
	Edged  WeaponType `yaml:"Edged,omitempty"`
}

/*
Impact
Pierce
Edged
Fire
etc.

---
Physical
Arcane
Spirit

*/
