package data

import (
	"reflect"

	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/imdario/mergo"
)

// MergeArchType is unique type for identifying how an archetype should be merged/added with another.
type MergeArchType uint8

// MergeArch provides an archetype ID and merge type pairing.
type MergeArch struct {
	ID   StringID
	Type MergeArchType
}

const (
	// ArchMerge determines that two archetypes should have their properties merged.
	ArchMerge MergeArchType = iota
	// ArchAdd determines that two archetypes should have their properties added together where applicable.
	ArchAdd
)

type RandomArchetype struct {
	Weight    float32   `yaml:"Weight,omitempty"`
	Archetype Archetype `yaml:"Archetype,omitempty"`
}

// ExitInfo represents the information used for simple exits and teleporters.
type ExitInfo struct {
	Name       string   `json:"Name" yaml:"Name,omitempty"`
	Y          *int     `json:"Y" yaml:"Y,omitempty"`
	X          *int     `json:"X" yaml:"X,omitempty"`
	Z          *int     `json:"Z" yaml:"Z,omitempty"`
	Touch      bool     `json:"Touch" yaml:"Touch,omitempty"`
	Cooldown   Duration `json:"Cooldown" yaml:"Cooldown,omitempty"`
	SizeRatio  float64  `json:"SizeRatio" yaml:"SizeRatio,omitempty"`
	Uses       int      `json:"Uses" yaml:"Uses,omitempty"`
	UniqueUses int      `json:"UniqueUses" yaml:"UniqueUses,omitempty"`
}

// Archetype represents a collection of data that should be used for the
// creation of Objects.
type Archetype struct {
	ArchID       StringID     `yaml:"-"` // Archetype ID used for generating objects and inheriting from.
	ArchIDs      []MergeArch  `yaml:"-"`
	Archs        []string     `json:"Archs" yaml:"Archs,omitempty"` // Archetypes to inherit from.
	ArchPointers []*Archetype `yaml:"-"`
	Arch         string       `json:"Arch" yaml:"Arch,omitempty"` // Archetype to inherit from. During post-parsing this is used to acquire and set the ArchID for inventory archetypes.
	SelfID       StringID     `yaml:"-"`                          // The Archetype's own SelfID
	Self         string       `json:"Self" yaml:"-"`
	Name         *string      `json:"Name" yaml:"Name,omitempty"` //
	//Name string
	Description *string             `json:"Description" yaml:"Description,omitempty"`
	Type        cdata.ArchetypeType `json:"Type" yaml:"Type,omitempty"`
	TypeHints   []string            `json:"TypeHints" yaml:"TypeHints,omitempty"`
	TypeHintIDs []StringID          `yaml:"-"`
	Anim        string              `json:"Anim" yaml:"Anim,omitempty"`
	AnimID      StringID            `yaml:"-"`
	Face        string              `json:"Face" yaml:"Face,omitempty"`
	FaceID      StringID            `yaml:"-"`
	Height      uint8               `json:"Height" yaml:"Height,omitempty"`
	Width       uint8               `json:"Width" yaml:"Width,omitempty"`
	Depth       uint8               `json:"Depth" yaml:"Depth,omitempty"`
	Matter      cdata.MatterType    `json:"Matter" yaml:"Matter,omitempty"`
	Blocking    cdata.MatterType    `json:"Blocking" yaml:"Blocking,omitempty"`
	//
	Audio      string   `json:"Audio" yaml:"Audio,omitempty"`
	AudioID    StringID `yaml:"-"`
	SoundSet   string   `json:"SoundSet" yaml:"SoundSet,omitempty"`
	SoundSetID StringID `yaml:"-"`
	SoundIndex int8     `json:"SoundIndex" yaml:"SoundIndex,omitempty"`
	// Lighting
	Light *Light `json:"Light" yaml:"Light,omitempty"`
	//
	Worth      *string             `json:"Worth" yaml:"Worth,omitempty"`
	Value      *string             `json:"Value" yaml:"Value,omitempty"`
	Count      *string             `json:"Count" yaml:"Count,omitempty"`
	Weight     *string             `json:"Weight" yaml:"Weight,omitempty"`
	Properties map[string]Variable `json:"Properties" yaml:"Properties,omitempty"`
	// Exit-related
	Exit *ExitInfo `json:"Exit" yaml:"Exit,omitempty"`
	//
	Timers []ArchetypeTimer `json:"Timers" yaml:"Timers,omitempty"`
	//
	Inventory []Archetype `json:"Inventory" yaml:"Inventory,omitempty"`
	// SkillTypes correspond to the skills used by weapons and similar.
	SkillTypes []SkillType `json:"SkillTypes" yaml:"SkillTypes,omitempty"`
	// CompetencyTypes are the competency types of a weapon or armor.
	CompetencyTypes []CompetencyType `json:"CompetencyTypes" yaml:"CompetencyTypes,omitempty"`
	// Skills are the skills levels contained by a character.
	Skills map[SkillType]Skill `json:"Skills" ts_type:"{[key:number]: Skill}" yaml:"Skills,omitempty"`
	// Competencies are the associated competencies' values that a character has.
	Competencies CompetenciesMap `json:"Competencies" ts_type:"{[key:number]: Competency}" yaml:"Competencies,omitempty"`
	// Resistances represents the attack type resistances of armor or a character.
	Resistances cdata.AttackTypes `json:"Resistances" ts_type:"{[key:number]: {[key:number]: number}}" yaml:"Resistances,omitempty"`
	// AttackTypes represents the attack types of a weapon or a character.
	AttackTypes cdata.AttackTypes `json:"AttackTypes" ts_type:"{[key:number]: {[key:number]:number}}" yaml:"AttackTypes,omitempty"`
	// Slots represents the slots information for equipment and characters.
	Slots Slots `json:"Slots" yaml:"Slots,omitempty"`
	// Reach represents how far this object can reach.
	Reach      uint8 `json:"Reach" yaml:"Reach,omitempty"`
	Attackable bool  `json:"Attackable" yaml:"Attackable,omitempty"`
	// Damage represents the damage of a weapon or otherwise.
	Damage *Damage `json:"Damage" yaml:"Damage,omitempty"`
	Armor  float64 `json:"Armor" yaml:"Armor,omitempty"`
	// Dodge represents an intrinsic dodge %. Only applicable to characters.
	Dodge        float64 `json:"Dodge" yaml:"Dodge,omitempty"`
	ChannelTime  uint16  `json:"ChannelTime" yaml:"ChannelTime,omitempty"`
	RecoveryTime uint16  `json:"RecoveryTime" yaml:"RecoveryTime,omitempty"`
	// Level represents the level of a skill or character.
	Level int `json:"Level" yaml:"Level,omitempty"`
	// Advancement represents the advancement of a skill or a character into the next level.
	Advancement float64 `json:"Advancement" yaml:"Advancement,omitempty"`
	// Efficiency represents the current efficiency of a skill.
	Efficiency float64 `json:"Efficiency" yaml:"Efficiency,omitempty"`
	//
	Attributes AttributeSets `json:"Attributes" yaml:"Attributes,omitempty"`
	// Hmm
	Statuses map[string]StatusMap `json:"Statuses" ts_type:"{[key:string]: {[key:string]: any}}" yaml:"Statuses,omitempty"`
	// Attitudes represent attitudes towards factions and similar.
	Attitudes Attitudes `json:"Attitudes" yaml:"Attitudes,omitempty"`
	Genera    string    `json:"Genera" yaml:"Genera,omitempty"`
	Species   string    `json:"Species" yaml:"Species,omitempty"`
	Factions  []string  `json:"Factions" yaml:"Factions,omitempty"`
	Legacy    string    `json:"Legacy" yaml:"Legacy,omitempty"`
	// Events are maps of EventIDs to EventResponses.
	Events *Events `json:"Events" yaml:"Events,omitempty"`
	//
	Specials Specials `json:"Specials" yaml:"Specials,omitempty"`
	//
	isCompiled   bool `yaml:"-"`
	isProcessing bool `yaml:"-"`
	isCompiling  bool `yaml:"-"`
}

type StatusMap map[string]interface{}

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

// IsCompiled returns if the archetype is compiled.
func (arch *Archetype) IsCompiled() bool {
	return arch.isCompiled
}

// SetCompiled sets the archetype to the given compiled state.
func (arch *Archetype) SetCompiled(b bool) {
	arch.isCompiled = b
}

// IsCompiling returns if the archetype is compiling.
func (arch *Archetype) IsCompiling() bool {
	return arch.isCompiling
}

// SetCompiling sets the archetype to the given compiling state.
func (arch *Archetype) SetCompiling(b bool) {
	arch.isCompiling = b
}

// Add adds properties from another archetype to itself, adding missing values and combining numerical values where able.
func (arch *Archetype) Add(other *Archetype) error {
	// Type hints are sent to the client during inspect or for inventory listings, so as to give hints as to what the item should be classified as.
	arch.TypeHintIDs = append(arch.TypeHintIDs, other.TypeHintIDs...)

	// Base values
	arch.Matter |= other.Matter
	arch.Blocking |= other.Blocking
	if arch.Worth == nil && other.Worth != nil {
		arch.Worth = &*other.Worth
	} else if other.Worth != nil {
		*arch.Worth += *other.Worth
	}
	if arch.Value == nil && other.Value != nil {
		arch.Value = &*other.Value
	} else if other.Value != nil {
		*arch.Value += *other.Value
	}
	if arch.Count == nil && other.Count != nil {
		arch.Count = &*other.Count
	} else if other.Count != nil {
		*arch.Count += *other.Count
	}
	if arch.Weight == nil && other.Weight != nil {
		arch.Weight = &*other.Weight
	} else if other.Weight != nil {
		*arch.Weight += *other.Weight
	}

	// Inventory
	for _, o := range other.Inventory {
		arch.Inventory = append(arch.Inventory, o)
	}

	// Skills and Competencies
	for k, v := range other.Skills {
		v2, ok := arch.Skills[k]
		if ok {
			v2.Experience += v.Experience
			arch.Skills[k] = v2
		} else {
			arch.Skills[k] = v
		}
	}
	for _, o := range other.SkillTypes {
		exists := false
		for _, s := range arch.SkillTypes {
			if o == s {
				exists = true
				break
			}
		}
		if !exists {
			arch.SkillTypes = append(arch.SkillTypes, o)
		}
	}
	for _, o := range other.CompetencyTypes {
		exists := false
		for _, s := range arch.CompetencyTypes {
			if o == s {
				exists = true
				break
			}
		}
		if !exists {
			arch.CompetencyTypes = append(arch.CompetencyTypes, o)
		}
	}
	for k, v := range other.Competencies {
		if v2, exists := arch.Competencies[k]; !exists {
			arch.Competencies[k] = v
		} else {
			v2.Efficiency += v.Efficiency
			arch.Competencies[k] = v2
		}
	}

	// Combat-related
	arch.Attackable = other.Attackable
	arch.Resistances.Add(other.Resistances)
	arch.AttackTypes.Add(other.AttackTypes)
	arch.Level += other.Level
	arch.Advancement += other.Advancement
	arch.Efficiency += other.Efficiency

	if arch.Damage == nil {
		arch.Damage = &Damage{}
	}

	if other.Damage != nil {
		arch.Damage.Add(other.Damage)
	}

	// Attributes
	arch.Attributes.Physical.Add(other.Attributes.Physical)
	arch.Attributes.Arcane.Add(other.Attributes.Arcane)
	arch.Attributes.Spirit.Add(other.Attributes.Spirit)

	// Slots. Note we're just using the resulting IDs, since add/merge is generally done after processing, which converts these to their appropriate IDs.
	arch.Slots.HasIDs = append(arch.Slots.HasIDs, other.Slots.HasIDs...)
	arch.Slots.GivesIDs = append(arch.Slots.GivesIDs, other.Slots.GivesIDs...)
	arch.Slots.UsesIDs = append(arch.Slots.UsesIDs, other.Slots.UsesIDs...)
	arch.Slots.Needs.MinIDs = append(arch.Slots.Needs.MinIDs, other.Slots.Needs.MinIDs...)
	arch.Slots.Needs.MaxIDs = append(arch.Slots.Needs.MaxIDs, other.Slots.Needs.MaxIDs...)

	// Brightness
	if arch.Light != nil {
		arch.Light.Add(other.Light)
	} else if other.Light != nil {
		arch.Light = &Light{}
		arch.Light.Add(other.Light)
	}

	// Exit-related logic
	if arch.Exit == nil && other.Exit != nil {
		y := *other.Exit.Y
		x := *other.Exit.X
		z := *other.Exit.Z
		arch.Exit = &ExitInfo{
			Name:      other.Exit.Name,
			Y:         &y,
			X:         &x,
			Z:         &z,
			Touch:     other.Exit.Touch,
			Cooldown:  other.Exit.Cooldown,
			SizeRatio: other.Exit.SizeRatio,
		}
	}

	// Factions
	for _, f := range other.Factions {
		exists := false
		for _, f2 := range arch.Factions {
			if f == f2 {
				exists = true
				break
			}
		}
		if !exists {
			arch.Factions = append(arch.Factions, f)
		}
	}

	return nil
}

// Merge will attempt to merge any missing properties from another archetype to this one.
func (arch *Archetype) Merge(other *Archetype) error {
	if err := mergo.Merge(arch, other); err != nil {
		return err
	}
	return nil
}

// GetField returns a reflect.Value of the target field in the archetype's properties.
func (arch *Archetype) GetField(fieldName string) reflect.Value {
	s := reflect.ValueOf(arch).Elem()
	f := s.FieldByName(fieldName)
	return f
}
