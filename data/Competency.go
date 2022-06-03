package data

import "fmt"

// Competency is the floating value of competency for a competency type. These range from 0 to 100.
type Competency struct {
	Efficiency float64 `json:"Efficiency" yaml:"Efficiency,omitempty"`
}

// CompetencyType is a type for a given competency.
type CompetencyType uint32

// CompetenciesMap is a convenience type mapping for our competencies.
type CompetenciesMap map[CompetencyType]Competency

// These are the various competency types within chimera.
const (
	//Incompetency CompetencyType = 0
	Incompetency CompetencyType = iota
	// Weapon styles
	//AxesCompetency = 1 << iota
	AxesCompetency
	HammersCompetency
	DaggersCompetency
	SwordsCompetency
	FlailsCompetency
	PolearmsCompetency
	// Fighting styles
	OneHandedCompetency
	TwoHandedCompetency
	DualHandedCompetency
	ShieldCompetency
	//
	PugilismCompetency
	PushDaggersCompetency
	// Ranged styles
	DrawnCompetency  // bows
	ThrownCompetency // throwing weapons
	AimedCompetency  // crossbows
	// Arcane/Spirit combat styles
	RayCompetency
	ConeCompetency
	ExplosionCompetency
	ChannelCompetency
	// Arcane types
	KineticCompetency
	TemperatureCompetency
	MaterializeCompetency
	// Spirit types
	HealCompetency
	HarmCompetency
	BlessCompetency
	CurseCompetency
	ProtectCompetency
	WeakenCompetency
	// Other
	DodgeCompetency
	LightArmorCompetency
	MediumArmorCompetency
	HeavyArmorCompetency
)

// StringToCompetencyMap maps strings to their corresponding CompetencyTypes.
var StringToCompetencyMap = map[string]CompetencyType{
	"Incompetency": Incompetency,
	"Axes":         AxesCompetency,
	"Hammers":      HammersCompetency,
	"Daggers":      DaggersCompetency,
	"Swords":       SwordsCompetency,
	"Flails":       FlailsCompetency,
	"Polearms":     PolearmsCompetency,
	//
	"One Handed":  OneHandedCompetency,
	"Two Handed":  TwoHandedCompetency,
	"Dual Handed": DualHandedCompetency,
	"Shield":      ShieldCompetency,
	//
	"Pugilism":     PugilismCompetency,
	"Push Daggers": PushDaggersCompetency,
	//
	"Drawn":  DrawnCompetency,
	"Thrown": ThrownCompetency,
	"Aimed":  AimedCompetency,
	//
	"Ray":       RayCompetency,
	"Cone":      ConeCompetency,
	"Explosion": ExplosionCompetency,
	"Channel":   ChannelCompetency,
	//
	"Kinetic":     KineticCompetency,
	"Temperature": TemperatureCompetency,
	"Materialize": MaterializeCompetency,
	//
	"Heal":    HealCompetency,
	"Harm":    HarmCompetency,
	"Bless":   BlessCompetency,
	"Curse":   CurseCompetency,
	"Protect": ProtectCompetency,
	"Weaken":  WeakenCompetency,
	//
	"Dodge":        DodgeCompetency,
	"Light Armor":  LightArmorCompetency,
	"Medium Armor": MediumArmorCompetency,
	"Heavy Armor":  HeavyArmorCompetency,
}

// CompetencyToStringMap maps CompetencyTypes to their corresponding string.
var CompetencyToStringMap = map[CompetencyType]string{
	// Melee
	Incompetency:       "Incompetency",
	AxesCompetency:     "Axes",
	HammersCompetency:  "Hammers",
	DaggersCompetency:  "Daggers",
	SwordsCompetency:   "Swords",
	FlailsCompetency:   "Flails",
	PolearmsCompetency: "Polearms",
	//
	OneHandedCompetency:  "One Handed",
	TwoHandedCompetency:  "Two Handed",
	DualHandedCompetency: "Dual Handed",
	ShieldCompetency:     "Shield",
	//
	PugilismCompetency:    "Pugilism",
	PushDaggersCompetency: "Push Daggers",
	//
	DrawnCompetency:  "Drawn",
	ThrownCompetency: "Thrown",
	AimedCompetency:  "Aimed",
	//
	RayCompetency:       "Ray",
	ConeCompetency:      "Cone",
	ExplosionCompetency: "Explosion",
	ChannelCompetency:   "Channel",
	//
	KineticCompetency:     "Kinetic",
	TemperatureCompetency: "Temperature",
	MaterializeCompetency: "Materialize",
	//
	HealCompetency:    "Heal",
	HarmCompetency:    "Harm",
	BlessCompetency:   "Bless",
	CurseCompetency:   "Curse",
	ProtectCompetency: "Protect",
	WeakenCompetency:  "Weaken",
	//
	DodgeCompetency:       "Dodge",
	LightArmorCompetency:  "Light Armor",
	MediumArmorCompetency: "Medium Armor",
	HeavyArmorCompetency:  "Heavy Armor",
}

// UnmarshalYAML unmarshals an ArchetypeType from a string.
func (ctype *CompetencyType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var value string
	if err := unmarshal(&value); err != nil {
		return err
	}
	if v, ok := StringToCompetencyMap[value]; ok {
		*ctype = v
		return nil
	}
	*ctype = Incompetency
	return fmt.Errorf("Unknown CompetencyType '%s'", value)
}

// MarshalYAML marshals an ArchetypeType into a string.
func (ctype CompetencyType) MarshalYAML() (interface{}, error) {
	if v, ok := CompetencyToStringMap[ctype]; ok {
		return v, nil
	}
	return "Incompetency", nil
}
