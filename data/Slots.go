package data

// Slots are used to define limits for equipment and characters.
type Slots struct {
	// Has represents how many slots the given character (or potentially slotted item) has.
	Has    []string   `json:"Has" yaml:"Has,omitempty"`
	HasIDs []StringID `yaml:"-"`
	// Needs are what slots are required for using the item.
	Needs SlotsNeeds `json:"Needs" yaml:"Needs,omitempty"`
	// Uses are the slots that the item takes up on its owner.
	Uses    []string   `json:"Uses" yaml:"Uses,omitempty"`
	UsesIDs []StringID `yaml:"-"`
	// Gives are the slots that the item gives to its owner.
	Gives    []string   `json:"Gives" yaml:"Gives,omitempty"`
	GivesIDs []StringID `yaml:"-"`
}

// SlotsNeeds are minimum and maximum slot requirements. These do not modify Has.
type SlotsNeeds struct {
	// Min defines the minimum amount of slots required.
	Min    []string   `json:"Min" yaml:"Min,omitempty"`
	MinIDs []StringID `yaml:"-"`
	// Max defines the maximum amount of slots required.
	Max    []string   `json:"Max" yaml:"Max,omitempty"`
	MaxIDs []StringID `yaml:"-"`
}
