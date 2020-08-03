package data

import (
	cdata "github.com/chimera-rpg/go-common/data"
)

// TODO: Should most Archetype properties be either StringExpression or NumberExpression? These would be string and int based types that can pull properties from their owning Archetype during Object creation. They likely should be a postfix-structured stack that can be passed some sort of context stack that contains the target Object and/or Archetype. We could also just have them as strings until they are instantized into an Object.

// Archetype represents a collection of data that should be used for the
// creation of Objects.
type Archetype struct {
	ArchID  StringID         `yaml:"-"` // Archetype ID used for generating objects and inheriting from.
	ArchIDs []StringID       `yaml:"-"`
	Archs   []string         `yaml:"Archs,omitempty"` // Archetypes to inherit from.
	Arch    string           `yaml:"Arch,omitempty"`  // Archetype to inherit from. During post-parsing this is used to acquire and set the ArchID for inventory archetypes.
	SelfID  StringID         `yaml:"-"`               // The Archetype's own SelfID
	Name    StringExpression `yaml:"Name,omitempty"`  // StringExpression
	//Name string
	Description StringExpression    `yaml:"Description,omitempty"` // StringExpression
	Type        cdata.ArchetypeType `yaml:"Type,omitempty"`
	Anim        string              `yaml:"Anim,omitempty"`
	AnimID      StringID            `yaml:"-"`
	Face        string              `yaml:"Face,omitempty"`
	FaceID      StringID            `yaml:"-"`
	Height      uint8               `yaml:"Height,omitempty"`
	Width       uint8               `yaml:"Width,omitempty"`
	Depth       uint8               `yaml:"Depth,omitempty"`
	Matter      cdata.MatterType    `yaml:"Matter,omitempty"`
	Blocking    cdata.MatterType    `yaml:"Blocking,omitempty"`
	//
	Value      StringExpression    `yaml:"Value,omitempty"`  // NumberExpression
	Count      StringExpression    `yaml:"Count,omitempty"`  // NumberExpression
	Weight     StringExpression    `yaml:"Weight,omitempty"` // NumberExpression
	Properties map[string]Variable `yaml:"Properties,omitempty"`
	//
	Inventory []Archetype `yaml:"Inventory,omitempty"`
	// Skills are the skills contained by a character.
	Skills []Archetype `yaml:"Skills,omitempty"`
	// SkillTypes correspond to the skills used by weapons and similar.
	SkillTypes []SkillType `yaml:"SkillTypes,omitempty"`
	// CompetencyTypes are the competency types of a weapon.
	CompetencyTypes []CompetencyType `yaml:"CompetencyTypes,omitempty"`
	// Competencies are the associated competencies' values that a skill instance has.
	Competencies map[CompetencyType]Competency `yaml:"Competencies,omitempty"`
	// CompetencyGroups are the competencies that a skill will train even if only one in a group is being trained.
	SharedCompetencyTypes [][]CompetencyType `yaml:"SharedCompetencyTypes,omitempty"`
	// Resistances represents the attack type resistances of armor or a character.
	Resistances AttackTypes `yaml:"Resistances,omitempty"`
	// AttackTypes represents the attack types of a weapon or a character.
	AttackTypes AttackTypes `yaml:"AttackTypes,omitempty"`
	// Level represents the level of a skill or character.
	Level int `yaml:"Level,omitempty"`
	// Advancement represents the advancement of a skill or a character into the next level.
	Advancement float32 `yaml:"Advancement,omitempty"`
	//
	Attributes struct {
		Physical Attributes `yaml:"Physical,omitempty"`
		Arcane   Attributes `yaml:"Arcane,omitempty"`
		Spirit   Attributes `yaml:"Spirit,omitempty"`
	} `yaml:"Attributes,omitempty"`
	//
	isCompiled bool `yaml:"-"`
}

// NewArchetype creates a new, blank archetype.
func NewArchetype() Archetype {
	return Archetype{
		Properties: make(map[string]Variable),
	}
}

// UnmarshalYAML unmarshals source YAML into an Archetype.
/*func (arch *Archetype) UnmarshalYAML(unmarshal func(interface{}) error) error {
	arch.Properties = make(map[string]Variable)

	kvs := make(map[string]interface{})
	err := unmarshal(&kvs)
	if err != nil {
		return err
	}
	if err := arch.fromMap(kvs); err != nil {
		return err
	}
	return nil
}*/

func (arch *Archetype) setProperty(key string, value interface{}) {
	switch v := value.(type) {
	case int:
		arch.Properties[key] = Int(v)
	case string:
		arch.Properties[key] = String(v)
	case bool:
		arch.Properties[key] = Bool(v)
	}
}

func (arch *Archetype) IsCompiled() bool {
	return arch.isCompiled
}

func (arch *Archetype) SetCompiled(b bool) {
	arch.isCompiled = b
}
