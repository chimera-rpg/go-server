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
	ArchetypePC
	ArchetypeNPC
	ArchetypeTile
	ArchetypeFloor
	ArchetypeWall
	ArchetypeItem
	ArchetypeBullet
	ArchetypeGeneric
)

// Archetype represents a collection of data that should be used for the
// creation of Objects.
type Archetype struct {
	Arch string // This value should always map to its place in game data's templates
	Name Variable
	//Name string
	Description Variable
	Type        ArchetypeType
	Anim        Variable // TODO: This should reference an already compiled AnimId
	//
	Value      Variable
	Count      Variable
	Properties map[string]Variable
	Inventory  map[string]Archetype
}

// NewArchetype creates a new, blank archetype.
func NewArchetype() Archetype {
	return Archetype{
		Properties: make(map[string]Variable),
		Inventory:  make(map[string]Archetype),
	}
}

func (arch *Archetype) setType(value string) error {
	switch value {
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

func (arch *Archetype) setStructProperty(key string, value string) error {
	switch key {
	case "Arch":
		arch.Arch = value
	case "Anim":
		arch.Anim = String(value)
	case "Description":
		arch.Description = String(value)
	case "Name":
		arch.Name = String(value)
	case "Type":
		arch.setType(value)
	case "Value":
		arch.Value = Expression(value)
	case "Count":
		arch.Count = Expression(value)
	default:
		arch.Properties[key] = Expression(value)
	}
	return nil
}
func (arch *Archetype) addProperty(key string, value string) error {
	arch.Properties[key] = String(value)
	return nil
}
