package world

import (
	"errors"

	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
)

// ObjectWeapon represents a skill.
type ObjectWeapon struct {
	Object
	name        string
	damaged     float32 // How damaged the weapon is.
	attackTypes data.AttackTypes
	// TODO: attack types
}

// NewObjectWeapon creates a skill object from the given archetype.
func NewObjectWeapon(a *data.Archetype) (o *ObjectWeapon) {
	o = &ObjectWeapon{
		Object:      Object{Archetype: a},
		attackTypes: a.AttackTypes,
	}

	//o.setArchetype(a)

	return
}

// getType returns the Archetype type.
func (o *ObjectWeapon) getType() cdata.ArchetypeType {
	return cdata.ArchetypeWeapon
}

func (o *ObjectWeapon) GetDamages(skills map[data.SkillType]data.Skill, competencies map[data.CompetencyType]data.Competency, attributes *data.AttributeSets) (damages []Damage, err error) {
	base := o.Archetype.Damage
	// Multiply by the weapon's skills
	totalSkill := 0.0
	totalSkillCount := 0
	for _, s := range o.Archetype.SkillTypes {
		v, ok := skills[s]
		if !ok {
			// No skill, we cannot process!
			return nil, errors.New("missing skill " + data.SkillTypeToStringMap[s])
		}
		// FIXME: This should according to leveling table, not "experience".
		totalSkill += v.Experience
		totalSkillCount++
	}
	totalSkill /= float64(totalSkillCount)

	// Get our competency float modifier.
	totalCompetency := 0.0
	totalCompetencyCount := 0
	for _, c := range o.Archetype.CompetencyTypes {
		v, ok := competencies[c]
		if !ok {
			// No competency, we cannot process!
			return nil, errors.New("missing competency " + data.CompetencyToStringMap[c])
		}
		totalCompetency += v.Efficiency
		totalCompetencyCount++
	}
	totalCompetency /= float64(totalCompetencyCount)
	totalCompetency = 0.5 + totalCompetency/2

	for k, a := range o.Archetype.AttackTypes {
		damage := Damage{
			AttackType: k,
			BaseDamage: base,
			Competency: totalCompetency,
			Skill:      totalSkill,
			//Value:      base * (d / 100) * (totalSkill * totalCompetency),
		}
		for k2, d := range a {
			damage.AttackStyles[k2] = d
		}
		damages = append(damages, damage)
	}

	return damages, nil
}
