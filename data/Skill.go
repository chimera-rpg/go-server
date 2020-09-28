package data

import "fmt"

type SkillType uint32

const (
	NoSkill     SkillType = 0
	MeleeCombat           = 1 << iota
	HandToHand
	RangedCombat
	SpiritSkill
	ArcaneSkill
)

var StringToSkillTypeMap = map[string]SkillType{
	"No Skill":      NoSkill,
	"Melee Combat":  MeleeCombat,
	"Hand-to-Hand":  HandToHand,
	"Ranged Combat": RangedCombat,
	"Spirit":        SpiritSkill,
	"Arcane":        ArcaneSkill,
}

var SkillTypeToStringMap = map[SkillType]string{
	NoSkill:      "No Skill",
	MeleeCombat:  "Melee Combat",
	HandToHand:   "Hand-to-Hand",
	RangedCombat: "Ranged Combat",
	SpiritSkill:  "Spirit",
	ArcaneSkill:  "Arcane",
}

// UnmarshalYAML unmarshals an SkillType from a string.
func (stype *SkillType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var value string
	if err := unmarshal(&value); err != nil {
		return err
	}
	if v, ok := StringToSkillTypeMap[value]; ok {
		*stype = v
		return nil
	}
	*stype = NoSkill
	return fmt.Errorf("Unknown SkillType '%s'", value)
}

// MarshalYAML marshals an ArchetypeType into a string.
func (stype SkillType) MarshalYAML() (interface{}, error) {
	if v, ok := SkillTypeToStringMap[stype]; ok {
		return v, nil
	}
	return "No Skill", nil
}
