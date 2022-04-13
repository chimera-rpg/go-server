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
	Name       string   `yaml:"Name,omitempty"`
	Y          *int     `yaml:"Y,omitempty"`
	X          *int     `yaml:"X,omitempty"`
	Z          *int     `yaml:"Z,omitempty"`
	Touch      bool     `yaml:"Touch,omitempty"`
	Cooldown   Duration `yaml:"Cooldown,omitempty"`
	SizeRatio  float64  `yaml:"SizeRatio,omitempty"`
	Uses       int      `yaml:"Uses,omitempty"`
	UniqueUses int      `yaml:"UniqueUses,omitempty"`
}

// Archetype represents a collection of data that should be used for the
// creation of Objects.
type Archetype struct {
	ArchID       StringID     `yaml:"-"` // Archetype ID used for generating objects and inheriting from.
	ArchIDs      []MergeArch  `yaml:"-"`
	Archs        []string     `yaml:"Archs,omitempty"` // Archetypes to inherit from.
	ArchPointers []*Archetype `yaml:"-"`
	Arch         string       `yaml:"Arch,omitempty"` // Archetype to inherit from. During post-parsing this is used to acquire and set the ArchID for inventory archetypes.
	SelfID       StringID     `yaml:"-"`              // The Archetype's own SelfID
	Name         *string      `yaml:"Name,omitempty"` //
	//Name string
	Description *string             `yaml:"Description,omitempty"`
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
	Audio      string   `yaml:"Audio,omitempty"`
	AudioID    StringID `yaml:"-"`
	SoundSet   string   `yaml:"SoundSet,omitempty"`
	SoundSetID StringID `yaml:"-"`
	SoundIndex int8     `yaml:"SoundIndex,omitempty"`
	//
	Worth      *string             `yaml:"Worth,omitempty"`
	Value      *string             `yaml:"Value,omitempty"`
	Count      *string             `yaml:"Count,omitempty"`
	Weight     *string             `yaml:"Weight,omitempty"`
	Properties map[string]Variable `yaml:"Properties,omitempty"`
	// Exit-related
	Exit *ExitInfo `yaml:"Exit,omitempty"`
	//
	Timers []ArchetypeTimer `yaml:"Timers,omitempty"`
	//
	Inventory []Archetype `yaml:"Inventory,omitempty"`
	// SkillTypes correspond to the skills used by weapons and similar.
	SkillTypes []SkillType `yaml:"SkillTypes,omitempty"`
	// CompetencyTypes are the competency types of a weapon or armor.
	CompetencyTypes []CompetencyType `yaml:"CompetencyTypes,omitempty"`
	// Skills are the skills levels contained by a character.
	Skills map[SkillType]Skill `yaml:"Skills,omitempty"`
	// Competencies are the associated competencies' values that a character has.
	Competencies map[CompetencyType]Competency `yaml:"Competencies,omitempty"`
	// Resistances represents the attack type resistances of armor or a character.
	Resistances cdata.AttackTypes `yaml:"Resistances,omitempty"`
	// AttackTypes represents the attack types of a weapon or a character.
	AttackTypes cdata.AttackTypes `yaml:"AttackTypes,omitempty"`
	// Reach represents how far this object can reach.
	Reach      uint8 `yaml:"Reach,omitempty"`
	Attackable bool  `yaml:"Attackable,omitempty"`
	// Damage represents the damage of a weapon or otherwise.
	Damage *Damage `yaml:"Damage,omitempty"`
	Armor  float64 `yaml:"Armor,omitempty"`
	// Dodge represents an intrinsic dodge %. Only applicable to characters.
	Dodge        float64 `yaml:"Dodge,omitempty"`
	ChannelTime  uint16  `yaml:"ChannelTime,omitempty"`
	RecoveryTime uint16  `yaml:"RecoveryTime,omitempty"`
	// Level represents the level of a skill or character.
	Level int `yaml:"Level,omitempty"`
	// Advancement represents the advancement of a skill or a character into the next level.
	Advancement float64 `yaml:"Advancement,omitempty"`
	// Efficiency represents the current efficiency of a skill.
	Efficiency float64 `yaml:"Efficiency,omitempty"`
	//
	Attributes AttributeSets `yaml:"Attributes,omitempty"`
	// Hmm
	Statuses map[string]map[string]interface{} `yaml:"Statuses,omitempty"`
	// Attitudes represent attitudes towards factions and similar.
	Attitudes Attitudes `yaml:"Attitudes,omitempty"`
	Genera    string    `yaml:"Genera,omitempty"`
	Family    string    `yaml:"Family,omitempty"`
	Species   string    `yaml:"Species,omitempty"`
	Factions  []string  `yaml:"Factions,omitemity"`
	Legacy    string    `yaml:"Legacy,omitempty"`
	// Events are maps of EventIDs to EventResponses.
	Events *Events `yaml:"Events,omitempty"`
	//
	isCompiled   bool `yaml:"-"`
	isProcessing bool `yaml:"-"`
	isCompiling  bool `yaml:"-"`
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

// IsCompiled returns if the archetype is compiled.
func (arch *Archetype) IsCompiled() bool {
	return arch.isCompiled
}

// SetCompiled sets the archetype to the given compiled state.
func (arch *Archetype) SetCompiled(b bool) {
	arch.isCompiled = b
}

// Add adds properties from another archetype to itself, adding missing values and combining numerical values where able.
func (arch *Archetype) Add(other *Archetype) error {
	arch.Matter |= other.Matter
	arch.Blocking |= other.Blocking
	arch.Attackable = other.Attackable
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

	for _, o := range other.Inventory {
		arch.Inventory = append(arch.Inventory, o)
	}
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

	arch.Resistances.Add(other.Resistances)
	arch.AttackTypes.Add(other.AttackTypes)
	arch.Level += other.Level
	arch.Advancement += other.Advancement
	arch.Efficiency += other.Efficiency

	if arch.Damage == nil {
		arch.Damage = &Damage{}
	}
	arch.Damage.Add(other.Damage)

	arch.Attributes.Physical.Add(other.Attributes.Physical)
	arch.Attributes.Arcane.Add(other.Attributes.Arcane)
	arch.Attributes.Spirit.Add(other.Attributes.Spirit)

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
