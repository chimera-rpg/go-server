package data

// Slots are used to define limits for equipment and characters.
type Slots struct {
	Has      []string   `json:"Has" yaml:"Has,omitempty"`
	HasIDs   []StringID `yaml:"-"`
	Needs    SlotsNeeds `json:"Needs" yaml:"Needs,omitempty"`
	Uses     []string   `json:"Uses" yaml:"Uses,omitempty"`
	UsesIDs  []StringID `yaml:"-"`
	Gives    []string   `json:"Gives" yaml:"Gives,omitempty"`
	GivesIDs []StringID `yaml:"-"`
}

// SlotsNeeds are minimum and maximum slot requirements. These do not modify Has.
type SlotsNeeds struct {
	Min    []string   `json:"Min" yaml:"Min,omitempty"`
	MinIDs []StringID `yaml:"-"`
	Max    []string   `json:"Max" yaml:"Max,omitempty"`
	MaxIDs []StringID `yaml:"-"`
}
