package world

import (
	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
	"time"
)

// ObjectPC represents player characters.
type ObjectPC struct {
	Object
	//
	name          string
	maxHp         int
	level         int
	race          string
	count         int
	value         int
	mapUpdateTime uint8 // Corresponds to the map's updateTime -- if they are out of sync then the player will sample its view space.
	resistance    AttackTypes
	abilityScores AbilityScores
}

// NewObjectPC creates a new ObjectPC from the given archetype.
func NewObjectPC(a *data.Archetype) (o *ObjectPC) {
	o = &ObjectPC{
		Object: Object{
			Archetype: a,
		},
	}

	//o.setArchetype(a)

	return
}

// NewObjectPCFromCharacter creates a new ObjectPC from the given character data.
func NewObjectPCFromCharacter(c *data.Character) (o *ObjectPC) {
	o = &ObjectPC{
		Object: Object{
			Archetype: &c.Archetype,
		},
		name: c.Name,
	}
	return
}

func (o *ObjectPC) setArchetype(targetArch *data.Archetype) {
	// First inherit from another Archetype if ArchID is set.
	/*mutatedArch := data.NewArchetype()
	for targetArch != nil {
		if err := mergo.Merge(&mutatedArch, targetArch); err != nil {
			log.Fatal("o no")
		}
		targetArch = targetArch.InheritArch
	}

	o.name, _ = mutatedArch.Name.GetString()*/
}

func (o *ObjectPC) update(delta time.Duration) {
	doTilesBlock := func(targetTiles []*Tile) bool {
		isBlocked := false
		matter := o.GetArchetype().Matter
		for _, tT := range targetTiles {
			for _, tO := range tT.objects {
				switch tO := tO.(type) {
				case *ObjectBlock:
					// Check if the target object blocks our matter.
					if tO.blocking.Is(matter) {
						isBlocked = true
					}
				}
			}
		}
		return isBlocked
	}

	// Add a falling timer if we've moved and should fall.
	var s *StatusFalling
	if o.hasMoved && !o.HasStatus(s) {
		m := o.tile.gameMap
		if m != nil {
			_, fallingTiles, err := m.GetObjectPartTiles(o, -1, 0, 0)

			if !doTilesBlock(fallingTiles) && err == nil {
				o.AddStatus(&StatusFalling{})
			}
		}
	}
	//
	o.Object.update(delta)
}

func (o *ObjectPC) getType() cdata.ArchetypeType {
	return cdata.ArchetypeNPC
}
