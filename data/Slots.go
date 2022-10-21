package data

type SlotsNeeds struct {
	Min    []string   `json:"Min" yaml:"Min,omitempty"`
	MinIDs []StringID `yaml:"-"`
	Max    []string   `json:"Max" yaml:"Max,omitempty"`
	MaxIDs []StringID `yaml:"-"`
}

type Slots struct {
	Has      []string   `json:"Has" yaml:"Has,omitempty"`
	HasIDs   []StringID `yaml:"-"`
	Needs    SlotsNeeds `json:"Needs" yaml:"Needs,omitempty"`
	Uses     []string   `json:"Uses" yaml:"Uses,omitempty"`
	UsesIDs  []StringID `yaml:"-"`
	Gives    []string   `json:"Gives" yaml:"Gives,omitempty"`
	GivesIDs []StringID `yaml:"-"`
}
