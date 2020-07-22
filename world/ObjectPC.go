package world

import (
	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
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
	canMove       bool
	fallTimer     int64
}

// NewObjectPC creates a new ObjectPC from the given archetype.
func NewObjectPC(a *data.Archetype) (o *ObjectPC) {
	o = &ObjectPC{
		Object: Object{
			Archetype: a,
		},
		canMove: true,
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
		name:    c.Name,
		canMove: true,
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

func (o *ObjectPC) update(delta int64) {
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

	// Handle if we are falling or should be falling.
	if o.state&FallingState != 0 {
		o.fallTimer += delta
		if o.fallTimer >= 18867 { // roughly 53 meters / second
			o.fallTimer -= 18867
			m := o.tile.gameMap
			if m != nil {
				_, fallingTiles, err := m.GetObjectPartTiles(o, -1, 0, 0)

				if doTilesBlock(fallingTiles) && err == nil {
					o.canMove = true
					o.state = o.state &^ FallingState
				} else {
					if _, err := m.MoveObject(o, -1, 0, 0); err != nil {
						o.canMove = true
					}
				}
			}
		}
	}
	if o.hasMoved {
		m := o.tile.gameMap
		if m != nil {
			_, fallingTiles, err := m.GetObjectPartTiles(o, -1, 0, 0)

			if !doTilesBlock(fallingTiles) && err == nil {
				o.fallTimer = 0
				o.state |= FallingState
				o.canMove = false
			}
		}
		o.hasMoved = false
	}
}

func (o *ObjectPC) getType() cdata.ArchetypeType {
	return cdata.ArchetypeNPC
}
