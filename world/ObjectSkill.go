package world

import (
	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
)

// ObjectSkill represents a skill.
type ObjectSkill struct {
	Object
	name string
	//level        uint8                                   // Current level of the skill. Determines bonuses.
	//advancement  float32                                 // Advancement into the next level.
	//efficiency   float32                                 // Efficiency % of the skill. Increases by use, decreases over time.
	//competencies map[data.CompetencyType]data.Competency // Associated competencies for this skill.
}

// NewObjectSkill creates a skill object from the given archetype.
func NewObjectSkill(a *data.Archetype) (o *ObjectSkill) {
	o = &ObjectSkill{
		Object: Object{Archetype: a},
	}

	//o.setArchetype(a)

	return
}

// getType returns the Archetype type.
func (o *ObjectSkill) getType() cdata.ArchetypeType {
	return cdata.ArchetypeSkill
}
