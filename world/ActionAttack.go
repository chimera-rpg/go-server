package world

import (
	"errors"
	"time"
)

// ActionAttack represents an action that is an attack.
type ActionAttack struct {
	Action
	Target  ID
	Y, X, Z int
	// Type int -- swing, pierce, etc.
}

// NewActionAttack returns an instantized version of ActionAttack
func NewActionAttack(y, x, z int, target ID, cost time.Duration) *ActionAttack {
	return &ActionAttack{
		Action: Action{
			channel:  cost / 4,
			recovery: cost - cost/4,
		},
		Y:      y,
		X:      x,
		Z:      z,
		Target: target,
	}
}

func (m *Map) HandleActionAttack(a *ActionAttack) error {
	if a.Target != 0 {
		o2 := m.world.GetObject(a.Target)
		if o2 == nil {
			return errors.New("attack request for missing object")
		}
		if o2.GetTile().GetMap() != m {
			return errors.New("Attack request for object in different map")
		}
		if o2.Attackable() {
			switch attacker := a.object.(type) {
			case *ObjectCharacter:
				if attacker.Attack(o2) {
					break
				}
			}
		}
	} else if a.Y != 0 || a.X != 0 || a.Z != 0 {
		h, w, d := a.object.GetDimensions()
		t := a.object.GetTile()
		tiles := m.ShootRay(float64(t.Y)+float64(h)/2, float64(t.X)+float64(w)/2, float64(t.Z)+float64(d)/2, float64(a.Y), float64(a.X), float64(a.Z), func(t *Tile) bool {
			return true
		})
		objs := getUniqueObjectsInTiles(tiles)
		// Ignore our own tile.
		for _, o := range objs {
			// Ignore ourself.
			if o == a.object {
				continue
			}
			if o.Attackable() {
				switch attacker := a.object.(type) {
				case *ObjectCharacter:
					if attacker.Attack(o) {
						break
					}
				}
			}
		}
	}
	return nil
}
