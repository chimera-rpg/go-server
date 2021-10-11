package data

// Character represents a player character.
type Character struct {
	Name      string    `yaml:"Name"`
	Archetype Archetype `yaml:"Archetype"`
	SaveInfo  SaveInfo  `yaml:"SaveInfo"`
}

// SaveInfo is the positional information for a saved character.
type SaveInfo struct {
	Map      string `yaml:"Map"`
	Y        int    `yaml:"Y"`
	X        int    `yaml:"X"`
	Z        int    `yaml:"Z"`
	Stamina  int    `yaml:"Stamina"`
	Statuses int    `yaml:"Statuses"`
}
