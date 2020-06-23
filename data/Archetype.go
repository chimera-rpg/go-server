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
	//
	Value      StringExpression    `yaml:"Value,omitempty"`  // NumberExpression
	Count      StringExpression    `yaml:"Count,omitempty"`  // NumberExpression
	Weight     StringExpression    `yaml:"Weight,omitempty"` // NumberExpression
	Properties map[string]Variable `yaml:"Properties,omitempty"`
	Inventory  []Archetype         `yaml:"Inventory,omitempty"`
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
