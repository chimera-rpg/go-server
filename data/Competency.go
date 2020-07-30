package data

import "fmt"

type CompetencyType uint32

const (
	Incompetency CompetencyType = 0
	// Weapon styles
	AxesCompetency = 1 << iota
	DaggersCompetency
	SwordsCompetency
	FlailsCompetency
	PolearmsCompetency
	// Fighting styles
	OneHandedCompetency
	TwoHandedCompetency
	DualHandedCompetency
	ShieldCompetency
)

var StringToCompetencyMap = map[string]CompetencyType{
	"Incompetency": Incompetency,
	"Axes":         AxesCompetency,
	"Daggers":      DaggersCompetency,
	"Swords":       SwordsCompetency,
	"Flails":       FlailsCompetency,
	"Polearms":     PolearmsCompetency,
	//
	"One Handed":  OneHandedCompetency,
	"Two Handed":  TwoHandedCompetency,
	"Dual Handed": DualHandedCompetency,
	"Shield":      ShieldCompetency,
}

var CompetencyToStringMap = map[CompetencyType]string{
	// Melee
	Incompetency:       "Incompetency",
	AxesCompetency:     "Axes",
	DaggersCompetency:  "Daggers",
	SwordsCompetency:   "Swords",
	FlailsCompetency:   "Flails",
	PolearmsCompetency: "Polearms",
	//
	OneHandedCompetency:  "One Handed",
	TwoHandedCompetency:  "Two Handed",
	DualHandedCompetency: "Dual Handed",
	ShieldCompetency:     "Shield",
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
