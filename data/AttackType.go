package data

// AttackType represents the % damage type that a weapon does or armor protects from.
type AttackType uint32

const (
	NoAttackType AttackType = 0
	Physical                = 1 << iota
	Arcane
	Spirit
	//
	Impact
	Pierce
	Edged
	//
	Flame
	Frost
	Lightning
	Corrosive
	Force
	//
	Heal
	Harm
)

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

type AttackTypes map[AttackType]float32

func (a *AttackTypes) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*a = make(AttackTypes)
	var value map[string]float32
	if err := unmarshal(&value); err != nil {
		return err
	}
	for k, v := range value {
		if ak, ok := StringToAttackTypeMap[k]; ok {
			map[AttackType]float32(*a)[ak] = v / 100
		}
	}
	return nil
}

func (a AttackTypes) MarshalYAML() (interface{}, error) {
	r := make(map[string]float32)
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
