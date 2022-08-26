package world

import (
	"fmt"
	"math"

	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
)

// ObjectWeapon represents a skill.
type ObjectWeapon struct {
	Object
	// TODO: attack types
}

// NewObjectWeapon creates a skill object from the given archetype.
func NewObjectWeapon(a *data.Archetype) (o *ObjectWeapon) {
	o = &ObjectWeapon{
		Object: Object{Archetype: a},
	}

	//o.setArchetype(a)

	return
}

// getType returns the Archetype type.
func (o *ObjectWeapon) getType() cdata.ArchetypeType {
	return cdata.ArchetypeWeapon
}

type MissingSkillError struct {
	skillType data.SkillType
}

func (e *MissingSkillError) Error() string {
	return fmt.Sprintf("missing skill \"%s\"", data.SkillTypeToStringMap[e.skillType])
}

type MissingCompetencyError struct {
	competencyType data.CompetencyType
}

func (e *MissingCompetencyError) Error() string {
	return fmt.Sprintf("missing competency \"%s\"", data.CompetencyToStringMap[e.competencyType])
}

func GetDamages(w *ObjectWeapon, c *ObjectCharacter) (damages Damages, err error) {
	base := 0
	if w.Archetype.Damage != nil {
		base = w.Archetype.Damage.Value
	}
	// Multiply by the weapon's skills
	totalSkill := 0.0
	totalSkillCount := 0
	for _, s := range w.Archetype.SkillTypes {
		v, ok := c.Archetype.Skills[s]
		if !ok {
			// No skill, we cannot process!
			return nil, &MissingSkillError{s}
		}
		totalSkill += math.Floor(v.Experience)
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
			return nil, &MissingCompetencyError{ct}
		}
		totalCompetency += v.Efficiency
		totalCompetencyCount++
	}
	totalCompetency /= float64(totalCompetencyCount)
	totalCompetency = 0.5 + totalCompetency/2

	for k, a := range w.Archetype.AttackTypes {
		damage := Damage{
			AttackType: k,
			Styles:     make(map[cdata.AttackStyle]float64),
			BaseDamage: float64(base),
		}

		// Calculate bonus damage.
		if bonus, ok := w.Archetype.Damage.AttributeBonus[k]; ok {
			for attrK, attrV := range bonus {
				if k == cdata.Physical {
					charAttr := c.Archetype.Attributes.Physical.GetAttribute(attrK)
					damage.AttributeDamage += float64(charAttr) * attrV
				} else if k == cdata.Arcane {
					charAttr := c.Archetype.Attributes.Arcane.GetAttribute(attrK)
					damage.AttributeDamage += float64(charAttr) * attrV
				} else if k == cdata.Spirit {
					charAttr := c.Archetype.Attributes.Spirit.GetAttribute(attrK)
					damage.AttributeDamage += float64(charAttr) * attrV
				}
			}
		}

		// Calculate attack style damage.
		for k2, d := range a {
			damage.Styles[k2] = d * (totalSkill * totalCompetency)
			damage.StyleTotal += damage.Styles[k2]
		}

		damages = append(damages, damage)
	}

	return damages, nil
}

func (o *ObjectWeapon) GetMundaneInfo(near bool) cdata.ObjectInfo {
	info := o.Object.GetMundaneInfo(near)
	// TODO: ???
	return info
}
