package data

import (
	"fmt"
	"log"
	//"strconv"
)

// ArchetypeType is our identifier for the different... archetype types.
type ArchetypeType int

// These are the our Archetype types.
const (
	ArchetypeUnknown ArchetypeType = iota
	ArchetypeGenus
	ArchetypeSpecies
	ArchetypePC
	ArchetypeNPC
	ArchetypeTile
	ArchetypeFloor
	ArchetypeWall
	ArchetypeItem
	ArchetypeBullet
	ArchetypeGeneric
)

// TODO: Should most Archetype properties be either StringExpression or NumberExpression? These would be string and int based types that can pull properties from their owning Archetype during Object creation. They likely should be a postfix-structured stack that can be passed some sort of context stack that contains the target Object and/or Archetype. We could also just have them as strings until they are instantized into an Object.

// Archetype represents a collection of data that should be used for the
// creation of Objects.
type Archetype struct {
	ArchID   StringID // Archetype ID used for generating objects and inheriting from.
	copyArch string   `yaml:"Arch"` // Archetype to inherit from. During post-parsing this is used to acquire and set the ArchID for inventory archetypes.
	Name     Variable // StringExpression
	//Name string
	Description Variable // StringExpression
	Type        ArchetypeType
	AnimID      StringID
	//
	Value      Variable            // NumberExpression
	Count      Variable            // NumberExpression
	Weight     Variable            // NumberExpression
	Properties map[string]Variable // ?? StringExpression ??
	Inventory  map[string]Archetype
}

// NewArchetype creates a new, blank archetype.
func NewArchetype() Archetype {
	return Archetype{
		Properties: make(map[string]Variable),
		Inventory:  make(map[string]Archetype),
	}
}

// UnmarshalYAML unmarshals source YAML into an Archetype.
func (arch *Archetype) UnmarshalYAML(unmarshal func(interface{}) error) error {
	arch.Properties = make(map[string]Variable)
	arch.Inventory = make(map[string]Archetype)

	kvs := make(map[string]interface{})
	err := unmarshal(&kvs)
	if err != nil {
		return err
	}
	if err := arch.fromMap(kvs); err != nil {
		return err
	}
	return nil
}

func (arch *Archetype) fromMap(kvs map[string]interface{}) error {
	for k, v := range kvs {
		switch k {
		case "Type":
			if s, ok := v.(string); !ok {
				log.Printf("Error, wrong type for %s\n", k)
			} else {
				if err := arch.setType(s); err != nil {
					log.Println(err)
				}
			}
		case "Name":
			if s, ok := v.(string); !ok {
				log.Printf("Error, wrong type for %s\n", k)
			} else {
				arch.Name = String(s)
			}
		case "Description":
			if s, ok := v.(string); !ok {
				log.Printf("Error, wrong type for %s\n", k)
			} else {
				arch.Description = String(s)
			}
		case "Value":
			if s, ok := v.(int); !ok {
				log.Printf("Error, wrong type for %s\n", k)
			} else {
				arch.Value = Expression(s)
			}
		case "Count":
			if s, ok := v.(int); !ok {
				log.Printf("Error, wrong type for %s\n", k)
			} else {
				arch.Count = Expression(s)
			}
		case "Weight":
			if s, ok := v.(int); !ok {
				log.Printf("Error, wrong type for %s\n", k)
			} else {
				arch.Weight = Expression(s)
			}
		case "Inventory":
			if archs, ok := v.([]interface{}); !ok {
				log.Printf("Error, wrong type for %s\n", k)
			} else {
				log.Printf("TODO: Somehow generate Inventory for %+v.", archs)
			}
		default:
			arch.setProperty(k, v)
		}
	}
	return nil
}

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

func (arch *Archetype) setType(value string) error {
	switch value {
	case "Genus":
		arch.Type = ArchetypeGenus
	case "Species":
		arch.Type = ArchetypeSpecies
	case "PC":
		arch.Type = ArchetypePC
	case "NPC":
		arch.Type = ArchetypeNPC
	case "Tile":
		arch.Type = ArchetypeTile
	case "Floor":
		arch.Type = ArchetypeFloor
	case "Wall":
		arch.Type = ArchetypeWall
	case "Item":
		arch.Type = ArchetypeItem
	case "Bullet":
		arch.Type = ArchetypeBullet
	case "Generic":
		arch.Type = ArchetypeGeneric
	default:
		arch.Type = ArchetypeUnknown
		return fmt.Errorf("Unknown Type '%s' for arch %s", value, arch.Name)
	}
	return nil
}
