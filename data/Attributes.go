package data

// Attribute is a numeric value that represents a character's base ability.
type Attribute uint8

// Attributes represent the attribute scores for skills, combat, and more.
type Attributes struct {
	// Might represents general strength. Used for damage.
	Might Attribute `yaml:"Might,omitempty"`
	// Prowess represents endurance. Used for hit/arcane/divine points.
	Prowess Attribute `yaml:"Prowess,omitempty"`
	// Focus represents accuracy. Used for criticals.
	Focus Attribute `yaml:"Focus,omitempty"`
	// Sense represents ability to sense things. Used primary for passive skills.
	Sense Attribute `yaml:"Sense,omitempty"`
	// Haste represents how quickly one can do things. Used for attack speed and movement speed.
	Haste Attribute `yaml:"Haste,omitempty"`
	// Reaction represents how well one can dodge. Used for dodging attacks.
	Reaction Attribute `yaml:"Reaction,omitempty"`
}
