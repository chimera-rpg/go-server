package data

import (
	"fmt"
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

func (atype *ArchetypeType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var value string
	if err := unmarshal(&value); err != nil {
		return err
	}
	switch value {
	case "Genus":
		*atype = ArchetypeGenus
	case "Species":
		*atype = ArchetypeSpecies
	case "PC":
		*atype = ArchetypePC
	case "NPC":
		*atype = ArchetypeNPC
	case "Tile":
		*atype = ArchetypeTile
	case "Floor":
		*atype = ArchetypeFloor
	case "Wall":
		*atype = ArchetypeWall
	case "Item":
		*atype = ArchetypeItem
	case "Bullet":
		*atype = ArchetypeBullet
	case "Generic":
		*atype = ArchetypeGeneric
	default:
		*atype = ArchetypeUnknown
		return fmt.Errorf("Unknown Type '%s'", value)
	}
	return nil
}

// TODO: Should most Archetype properties be either StringExpression or NumberExpression? These would be string and int based types that can pull properties from their owning Archetype during Object creation. They likely should be a postfix-structured stack that can be passed some sort of context stack that contains the target Object and/or Archetype. We could also just have them as strings until they are instantized into an Object.

// Archetype represents a collection of data that should be used for the
// creation of Objects.
type Archetype struct {
	ArchID      StringID         `yaml:"-"`              // Archetype ID used for generating objects and inheriting from.
	Arch        string           `yaml:"Arch,omitempty"` // Archetype to inherit from. During post-parsing this is used to acquire and set the ArchID for inventory archetypes.
	InheritArch *Archetype       `yaml:"-"`
	SelfID      StringID         `yaml:"-"`              // The Archetype's own SelfID
	Name        StringExpression `yaml:"Name,omitempty"` // StringExpression
	//Name string
	Description StringExpression `yaml:"Description,omitempty"` // StringExpression
	Type        ArchetypeType    `yaml:"Type,omitempty"`
	Anim        string           `yaml:"Anim,omitempty"`
	AnimID      StringID         `yaml:"-"`
	Face        string           `yaml:"Face,omitempty"`
	FaceID      StringID         `yaml:"-"`
	//
	Value      StringExpression    `yaml:"Value,omitempty"`  // NumberExpression
	Count      StringExpression    `yaml:"Count,omitempty"`  // NumberExpression
	Weight     StringExpression    `yaml:"Weight,omitempty"` // NumberExpression
	Properties map[string]Variable `yaml:"Properties,omitempty"`
	Inventory  []Archetype         `yaml:"Inventory,omitempty"`
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

/*func (arch *Archetype) fromMap(kvs map[string]interface{}) error {
	for k, v := range kvs {
		switch k {
		case "Arch":
			if s, ok := v.(string); !ok {
				log.Printf("Error, wrong type for %s\n", k)
			} else {
				arch.Arch = s
			}
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
				arch.Name = BuildStringExpression(s)
			}
		case "Description":
			if s, ok := v.(string); !ok {
				log.Printf("Error, wrong type for %s\n", k)
			} else {
				arch.Description = BuildStringExpression(s)
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
			log.Printf("WOW: %T\n", v)
			if items, ok := v.([]interface{}); !ok {
				log.Printf("Error, wrong type for %s\n", k)
				log.Printf("%+v\n", v)
			} else {
				fmt.Printf("archs: %+v\n", items)
				for _, item := range items {
					fmt.Printf("FUCK: %T %+v\n", item, item)
					for key, value := range item.(map[string]interface{}) {
						fmt.Printf("DANG: %+v %+v\n", key, value)
					}
					//itemMap := make(map[string]interface{})
					/*for k, v := range items[i] {
						strKey := fmt.Sprintf("%v", k)
						itemMap[strKey] = v
					}
					itemArch.fromMap(itemMap)
					//var itemArch Archetype
					//arch.Inventory = append(arch.Inventory, itemArch)
				}
			}
		default:
			arch.setProperty(k, v)
		}
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
