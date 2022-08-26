package data

import "time"

// Character represents a player character.
type Character struct {
	Archetype Archetype `yaml:"Archetype"`
	SaveInfo  SaveInfo  `yaml:"SaveInfo"`
}

// SaveInfo is the positional information for a saved character.
type SaveInfo struct {
	Map      string    `yaml:"Map"`
	Y        int       `yaml:"Y"`
	X        int       `yaml:"X"`
	Z        int       `yaml:"Z"`
	HavenMap string    `yaml:"HavenMap"`
	HavenY   int       `yaml:"HavenY"`
	HavenX   int       `yaml:"HavenX"`
	HavenZ   int       `yaml:"HavenZ"`
	Time     time.Time `yaml:"Time"`
	Stamina  int       `yaml:"Stamina"`
	Statuses int       `yaml:"Statuses"`
}
