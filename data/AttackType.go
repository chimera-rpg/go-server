package data

// AttackType represents the % damage type that a weapon does or armor protects from.
type AttackType uint32

const (
	// NoAttackType should not occur.
	NoAttackType AttackType = 0
	// Physical represents physical attacks.
	Physical = 1 << iota
	// Arcane represents arcane attacks.
	Arcane
	// Spirit represents spirit attacks.
	Spirit
	// Impact represents physical attack types of a bludgeoning nature.
	Impact
	// Pierce represents physical attack types of a sharp, thrusting nature.
	Pierce
	// Edged represents physical attack types of an edged nature, such as swords.
	Edged
	// Flame represents fire.
	Flame
	// Frost represents cold.
	Frost
	// Lightning represents Zeus's mighty power.
	Lightning
	// Corrosive represents acid.
	Corrosive
	// Force represents mere force.
	Force
	// Heal does healing.
	Heal
	// Harm does harming.
	Harm
)

// StringToAttackTypeMap is the map of strings to attack types.
var StringToAttackTypeMap = map[string]AttackType{
	"Physical": Physical,
	"Arcane":   Arcane,
	"Spirit":   Spirit,

	"Impact": Impact,
	"Pierce": Pierce,
	"Edged":  Edged,

	"Flame":     Flame,
	"Frost":     Frost,
	"Lightning": Lightning,
	"Corrosive": Corrosive,
	"Force":     Force,

	"Heal": Heal,
	"Harm": Harm,
}

// AttackTypeToStringMap is a map of attack types to strings.
var AttackTypeToStringMap = map[AttackType]string{
	Physical: "Physical",
	Arcane:   "Arcane",
	Spirit:   "Spirit",

	Impact: "Impact",
	Pierce: "Pierce",
	Edged:  "Edged",

	Flame:     "Flame",
	Frost:     "Frost",
	Lightning: "Lightning",
	Corrosive: "Corrosive",
	Force:     "Force",

	Heal: "Heal",
	Harm: "Harm",
}

// AttackTypes is a map of AttackTypes to floats.
type AttackTypes map[AttackType]float64

// UnmarshalYAML unmarshals, converting attack type strings into AttackTypes.
func (a *AttackTypes) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*a = make(AttackTypes)
	var value map[string]float64
	if err := unmarshal(&value); err != nil {
		return err
	}
	for k, v := range value {
		if ak, ok := StringToAttackTypeMap[k]; ok {
			map[AttackType]float64(*a)[ak] = v / 100
		}
	}
	return nil
}

// MarshalYAML marshals, converting AttackTypes into strings.
func (a AttackTypes) MarshalYAML() (interface{}, error) {
	r := make(map[string]float64)
	for k, v := range a {
		if sk, ok := AttackTypeToStringMap[k]; ok {
			r[sk] = v * 100
		}
	}
	return r, nil
}

// Add adds missing attack types from another AttackTypes objects and combines any existing values.
func (a AttackTypes) Add(o AttackTypes) {
	for k, v := range o {
		if _, exists := a[k]; !exists {
			a[k] = v
		} else {
			a[k] += v
		}
	}
}

/*
Impact
Pierce
Edged
Fire
etc.

---
Physical
Arcane
Spirit

*/
