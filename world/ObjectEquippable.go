package world

import (
	"fmt"
	"math"

	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
)

type ObjectEquippable struct {
	Object
	armors         Armors
	finalArmors    Armors
	changedArmors  bool
	damages        Damages
	finalDamages   Damages
	changedDamages bool
}

func NewObjectEquippable(a *data.Archetype) (o *ObjectEquippable) {
	o = &ObjectEquippable{
		Object:         Object{Archetype: a},
		changedArmors:  true,
		changedDamages: true,
	}

	return
}

func (o *ObjectEquippable) Slots() map[data.StringID]int {
	return o.Archetype.Slots.HasIDs
}

// CalculateDamages calculates the weapon's base damages.
func (o *ObjectEquippable) CalculateDamages() {
	o.damages = Damages{}
	o.finalDamages = Damages{}
	if o.Archetype.Damage == nil || o.Archetype.Damage.Value == 0 {
		return
	}
	for k, a := range o.Archetype.AttackTypes {
		damage := Damage{
			AttackType: k,
			Styles:     make(map[cdata.AttackStyle]float64),
			BaseDamage: float64(o.Archetype.Damage.Value),
		}

		// Calculate attack style damage.
		for k2, d := range a {
			damage.Styles[k2] = d
			damage.StyleTotal += damage.Styles[k2]
		}

		o.damages = append(o.damages, damage)
	}
}

// GetDamages returns the adjusted Damages for the given character's skills, competencies, and attributes.
func (o *ObjectEquippable) GetDamages(c *ObjectCharacter) (damages Damages, err error) {
	if !o.changedDamages {
		return o.finalDamages, nil
	}
	// Multiply by the weapon's skills
	totalSkill := 0.0
	totalSkillCount := 0
	for _, s := range o.Archetype.SkillTypes {
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
	for _, ct := range o.Archetype.CompetencyTypes {
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

	for _, damage := range o.damages {
		// Calculate bonus damage.
		if bonus, ok := o.Archetype.Damage.AttributeBonus[damage.AttackType]; ok {
			for attrK, attrV := range bonus {
				if damage.AttackType == cdata.Physical {
					charAttr := c.Archetype.Attributes.Physical.GetAttribute(attrK)
					damage.AttributeDamage += float64(charAttr) * attrV
				} else if damage.AttackType == cdata.Arcane {
					charAttr := c.Archetype.Attributes.Arcane.GetAttribute(attrK)
					damage.AttributeDamage += float64(charAttr) * attrV
				} else if damage.AttackType == cdata.Spirit {
					charAttr := c.Archetype.Attributes.Spirit.GetAttribute(attrK)
					damage.AttributeDamage += float64(charAttr) * attrV
				}
			}
		}
		// Calculate attack style damage.
		for k, _ := range damage.Styles {
			damage.Styles[k] *= (totalSkill * totalCompetency)
			damage.StyleTotal += damage.Styles[k]
		}
		damages = append(damages, damage)
	}
	o.finalDamages = damages
	o.changedDamages = false
	return damages, nil
}

// CalculateArmors calculates the object's base armor values.
func (o *ObjectEquippable) CalculateArmors() {
	o.armors = Armors{}
	o.finalArmors = Armors{}
	if o.Archetype.Armor == 0 {
		return
	}
	var armors Armors
	for k, a := range o.Archetype.Resistances {
		armor := Armor{
			ArmorType: k,
			Styles:    make(map[cdata.AttackStyle]float64),
		}

		for k2, s := range a {
			armor.Styles[k2] = o.Archetype.Armor * s
		}
		armors = append(armors, armor)
	}

	o.armors = armors
}

// GetArmors returns the final calculated amount of armor using the given character's skills and competencies.
func (o *ObjectEquippable) GetArmors(c *ObjectCharacter) (armors Armors, err error) {
	if !o.changedArmors {
		return o.finalArmors, nil
	}
	// Multiply by the armors's skills
	totalSkill := 0.0
	totalSkillCount := 0
	for _, s := range o.Archetype.SkillTypes {
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
	for _, ct := range o.Archetype.CompetencyTypes {
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

	for _, a := range o.armors {
		for k := range a.Styles {
			a.Styles[k] *= (totalSkill * totalCompetency)
		}
		armors = append(armors, a)
	}
	o.finalArmors = armors
	o.changedArmors = false
	return armors, nil
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

func GetDamages(w *ObjectEquippable, c *ObjectCharacter) (damages Damages, err error) {
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

func GetArmors(a *ObjectEquippable, c *ObjectCharacter) (armors Armors, err error) {
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
