package world

import (
	"math"

	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
)

// ObjectArmor represents a skill.
type ObjectArmor struct {
	Object
	name        string
	damaged     float32 // How damaged the armor is.
	resistances cdata.AttackTypes
}

// NewObjectArmor creates a skill object from the given archetype.
func NewObjectArmor(a *data.Archetype) (o *ObjectArmor) {
	o = &ObjectArmor{
		Object: Object{Archetype: a},
	}

	//o.setArchetype(a)

	return
}

// getType returns the Archetype type.
func (o *ObjectArmor) getType() cdata.ArchetypeType {
	return cdata.ArchetypeArmor
}

func GetArmors(a *ObjectArmor, c *ObjectCharacter) (armors Armors, err error) {
	base := a.Archetype.Armor
	// Multiply by the armors's skills
	totalSkill := 0.0
	totalSkillCount := 0
	for _, s := range a.Archetype.SkillTypes {
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
	for _, ct := range a.Archetype.CompetencyTypes {
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

	for k, a := range a.Archetype.Resistances {
		armor := Armor{
			ArmorType: k,
			Styles:    make(map[cdata.AttackStyle]float64),
		}

		for k2, s := range a {
			armor.Styles[k2] = base * s * (totalSkill * totalCompetency)
		}
		armors = append(armors, armor)
	}

	return
}
