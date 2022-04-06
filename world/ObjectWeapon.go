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

func GetDamages(w *ObjectWeapon, c *ObjectCharacter) (damages []Damage, err error) {
	base := w.Archetype.Damage
	// Multiply by the weapon's skills
	totalSkill := 0.0
	totalSkillCount := 0
	for _, s := range w.Archetype.SkillTypes {
		v, ok := c.Archetype.Skills[s]
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
	for _, ct := range w.Archetype.CompetencyTypes {
		v, ok := c.Archetype.Competencies[ct]
		if !ok {
			// No competency, we cannot process!
			return nil, errors.New("missing competency " + data.CompetencyToStringMap[ct])
		}
		totalCompetency += v.Efficiency
		totalCompetencyCount++
	}
	totalCompetency /= float64(totalCompetencyCount)
	totalCompetency = 0.5 + totalCompetency/2

	for k, a := range w.Archetype.AttackTypes {
		damage := Damage{
			AttackType: k,
			//BaseDamage: float64(base.Value),
			//Competency: totalCompetency,
			//Skill:      totalSkill,
			//Value:      base * (d / 100) * (totalSkill * totalCompetency),
		}

		// Calculate bonus damage.
		if bonus, ok := w.Archetype.Damage.AttributeBonus[k]; ok {
			for attrK, attrV := range bonus {
				if k == data.Physical {
					charAttr := c.Archetype.Attributes.Physical.GetAttribute(attrK)
					damage.AttributeDamage += float64(charAttr) * attrV
				} else if k == data.Arcane {
					charAttr := c.Archetype.Attributes.Arcane.GetAttribute(attrK)
					damage.AttributeDamage += float64(charAttr) * attrV
				} else if k == data.Spirit {
					charAttr := c.Archetype.Attributes.Spirit.GetAttribute(attrK)
					damage.AttributeDamage += float64(charAttr) * attrV
				}
			}
		}

		// Calculate attack style damage.
		for k2, d := range a {
			damage.StyleDamages[k2] = float64(base.Value) * d * (totalSkill * totalCompetency)
		}

		damages = append(damages, damage)
	}

	return damages, nil
}
