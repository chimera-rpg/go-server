package data

// Armor represents the armor of a player, equipment, or otherwise as a percentile.
type Armor float32

// Armors is our collection of armor types.
type Armors struct {
	// Overarching armor types.
	Physical Armor `yaml:"Physical,omitempty"`
	Arcane   Armor `yaml:"Arcane,omitempty"`
	Spirit   Armor `yaml:"Spirit,omitempty"`
	// Subtypes.
	// TODO: Place armor types for whatever actual damage types exist.
}
