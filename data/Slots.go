package data

// Slots are used to define limits for equipment and characters.
type Slots struct {
	// Has represents how many slots the given character (or potentially slotted item) has.
	Has    map[string]int   `json:"Has" yaml:"Has,omitempty"`
	HasIDs map[StringID]int `yaml:"-"`
	// Needs are what slots are required for using the item.
	Needs SlotsNeeds `json:"Needs" yaml:"Needs,omitempty"`
	// Uses are the slots that the item takes up on its owner.
	Uses    map[string]int   `json:"Uses" yaml:"Uses,omitempty"`
	UsesIDs map[StringID]int `yaml:"-"`
	// Free are the slots that are available on the live character. These _should not_ be defined directly on the archetype files, as they are used for serialization of characters.
	Free    map[string]int   `json:"Free" yaml:"Free,omitempty"`
	FreeIDs map[StringID]int `yaml:"-"`
	// Gives are the slots that the item gives to its owner.
	Gives    map[string]int   `json:"Gives" yaml:"Gives,omitempty"`
	GivesIDs map[StringID]int `yaml:"-"`
}

// SlotsNeeds are minimum and maximum slot requirements. These do not modify Has.
type SlotsNeeds struct {
	// Min defines the minimum amount of slots required.
	Min    map[string]int   `json:"Min" yaml:"Min,omitempty"`
	MinIDs map[StringID]int `yaml:"-"`
	// Max defines the maximum amount of slots required.
	Max    map[string]int   `json:"Max" yaml:"Max,omitempty"`
	MaxIDs map[StringID]int `yaml:"-"`
}
