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
)

// StringToAttackTypeMap is the map of strings to attack types.
var StringToAttackTypeMap = map[string]AttackType{
	"Physical": Physical,
	"Arcane":   Arcane,
	"Spirit":   Spirit,
}

// AttackTypeToStringMap is a map of attack types to strings.
var AttackTypeToStringMap = map[AttackType]string{
	Physical: "Physical",
	Arcane:   "Arcane",
	Spirit:   "Spirit",
}

// AttackTypes is a map of AttackTypes to floats.
type AttackTypes map[AttackType]AttackStyles

type AttackStyles map[AttackStyle]float64

// UnmarshalYAML unmarshals, converting attack type strings into AttackTypes.
func (a *AttackTypes) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*a = make(AttackTypes)
	var value map[string]map[string]float64
	if err := unmarshal(&value); err != nil {
		return err
	}
	for k, v := range value {
		if ak, ok := StringToAttackTypeMap[k]; ok {
			(*a)[ak] = make(map[AttackStyle]float64)
			for k2, v2 := range v {
				if sk, ok := StringToAttackStyleMap[k2]; ok {
					(*a)[ak][sk] = v2 / 100
				}
			}
		}
	}
	return nil
}

// MarshalYAML marshals, converting AttackTypes into strings.
func (a AttackTypes) MarshalYAML() (interface{}, error) {
	r := make(map[string]map[string]float64)
	for k, v := range a {
		if sk, ok := AttackTypeToStringMap[k]; ok {
			r[sk] = make(map[string]float64)
			for k2, v2 := range v {
				if sk2, ok := AttackStyleToStringMap[k2]; ok {
					r[sk][sk2] = v2 * 100
				}
			}
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
			for k2, v2 := range v {
				if _, exists := a[k][k2]; !exists {
					a[k][k2] = v2
				} else {
					a[k][k2] += v2
				}
			}
		}
	}
}
