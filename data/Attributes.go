package data

// AttributesType is a numerical type for specific attribute sets.
type AttributesType uint8

// AttributesType consts
const (
	UnknownAttributes  AttributeType = 0
	PhysicalAttributes               = 1 << iota
	ArcaneAttributes
	SpiritAttributes
)

// AttributeSets is our set of a Attributes.
type AttributeSets struct {
	Physical Attributes `json:"Physical" yaml:"Physical,omitempty"`
	Arcane   Attributes `json:"Arcane" yaml:"Arcane,omitempty"`
	Spirit   Attributes `json:"Spirit" yaml:"Spirit,omitempty"`
}

// Add adds the physical, arcane, and spirits attributes to this attribute set.
func (as *AttributeSets) Add(o AttributeSets) {
	as.Physical.Add(o.Physical)
	as.Arcane.Add(o.Arcane)
	as.Spirit.Add(o.Spirit)
}

// AttributeValue is a numeric value that represents a character's base ability.
type AttributeValue int

// AttributeType is a numerical reference to specific attributes.
type AttributeType uint8

// AttributeType consts
const (
	UnknownAttribute AttributeType = 0
	Might                          = 1 << iota
	Prowess
	Focus
	Sense
	Haste
	Reaction
)

// StringToAttributeTypeMap is the map of strings to attributes.
var StringToAttributeTypeMap = map[string]AttributeType{
	"Might":    Might,
	"Prowess":  Prowess,
	"Focus":    Focus,
	"Sense":    Sense,
	"Haste":    Haste,
	"Reaction": Reaction,
}

// AttributeTypeToStringMap is the map of attributes to strings.
var AttributeTypeToStringMap = map[AttributeType]string{
	Might:    "Might",
	Prowess:  "Prowess",
	Focus:    "Focus",
	Sense:    "Sense",
	Haste:    "Haste",
	Reaction: "Reaction",
}

// Attributes represent the attribute scores for skills, combat, and more.
type Attributes struct {
	// Might represents general strength. Used for damage.
	Might AttributeValue `json:"Might" yaml:"Might,omitempty"`
	// Prowess represents endurance. Used for hit/arcane/divine points.
	Prowess AttributeValue `json:"Prowess" yaml:"Prowess,omitempty"`
	// Focus represents accuracy. Used for criticals.
	Focus AttributeValue `json:"Focus" yaml:"Focus,omitempty"`
	// Sense represents ability to sense things. Used primary for passive skills.
	Sense AttributeValue `json:"Sense" yaml:"Sense,omitempty"`
	// Haste represents how quickly one can do things. Used for attack speed and movement speed.
	Haste AttributeValue `json:"Haste" yaml:"Haste,omitempty"`
	// Reaction represents how well one can dodge. Used for dodging attacks.
	Reaction AttributeValue `json:"Reaction" yaml:"Reaction,omitempty"`
}

// Add adds together all attributes from another Attributes object.
func (a *Attributes) Add(o Attributes) {
	a.Might += o.Might
	a.Prowess += o.Prowess
	a.Focus += o.Focus
	a.Sense += o.Sense
	a.Haste += o.Haste
	a.Reaction += o.Reaction
}

// GetSpeedBonus gets the speed bonus for these attributes.
func (a *Attributes) GetSpeedBonus() AttributeValue {
	return a.Haste*5 + a.Reaction/4*5 + a.Might*2
}

// GetHealthBonus gets the health bonus for these attributes.
func (a *Attributes) GetHealthBonus() AttributeValue {
	return a.Prowess*8 + a.Might*2
}

// GetInspectBonus gets the inspect bonus for these attributes.
func (a *Attributes) GetInspectBonus() AttributeValue {
	return a.Sense + a.Haste/2 + a.Reaction/4
}

// GetAttributes gets the target attribute.
func (a *Attributes) GetAttribute(w AttributeType) AttributeValue {
	switch w {
	case Might:
		return a.Might
	case Prowess:
		return a.Prowess
	case Focus:
		return a.Focus
	case Sense:
		return a.Sense
	case Haste:
		return a.Haste
	case Reaction:
		return a.Reaction
	}
	return 0
}
