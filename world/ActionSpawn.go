package world

import (
	"github.com/chimera-rpg/go-server/data"
	log "github.com/sirupsen/logrus"
)

// ActionSpawn represents an action to spawn an object.
type ActionSpawn struct {
	Action
	Y, X, Z int  // coordinates to attempt spawning at
	Force   bool // If the spawn should be forced, ignoring any blocking/matter rules.
	Spawn   *data.SpawnArchetype
}

// HandleActionSpawn handles the spawning action.
func HandleActionSpawn(m *Map, a *ActionSpawn) error {
	if m == nil || a == nil {
		return nil
	}
	spawnItem := a.Spawn

	placedCoords := make(map[[3]int]struct{})
	count := spawnItem.Count.Random()

	spawn := func(y, x, z int) int {
		if object, err := m.world.CreateObjectFromArch(spawnItem.Archetype); err != nil {
			log.Warn("Spawn", err)
			return 2
		} else {
			if err := m.PlaceObject(object, y, x, z); err != nil {
				log.Warn("Spawn", err)
				return 1
			} else {
				placedCoords[[3]int{y, x, z}] = struct{}{}
				object.ResolveEvent(EventBirth{})
			}
		}
		return 0
	}

	if a.Force {
		spawn(a.Y+spawnItem.Placement.Y.Random(), a.X+spawnItem.Placement.X.Random(), a.Z+spawnItem.Placement.Z.Random())
		return nil
	}

	for i := 0; i < count; i++ {
		failed := false
		for i := -1; i < spawnItem.Retry; i++ {
			x := a.X + spawnItem.Placement.X.Random()
			y := a.Y + spawnItem.Placement.Y.Random()
			z := a.Z + spawnItem.Placement.Z.Random()

			// Deny spawning at same coord
			if y == a.Y && x == a.X && z == a.Z {
				continue
			}

			// Bail if overlap is false and we've already spawned at this location.
			if !spawnItem.Placement.Overlap {
				exists := false
				if _, ok := placedCoords[[3]int{y, x, z}]; ok {
					exists = true
				}
				if exists {
					continue
				}
			}

			// Check if the surface is ideal for us.
			h := int(spawnItem.Archetype.Height)
			w := int(spawnItem.Archetype.Width)
			d := int(spawnItem.Archetype.Depth)
			if h == 0 {
				h = 1
			}
			if w == 0 {
				w = 1
			}
			if d == 0 {
				d = 1
			}
			checkMatter := func(y, x, z int, matter *data.MatterType, checkMatter bool) bool {
				for yi := y; yi < y+h; yi++ {
					for xi := x; xi < x+w; xi++ {
						for zi := z; zi < z+d; zi++ {
							t2 := m.GetTile(yi, xi, zi)
							if t2 == nil {
								return false
							}
							if checkMatter {
								if t2.matter == *matter || t2.matter.Is(*matter) {
									return true
								}
							} else {
								if t2.blocking == *matter || t2.blocking.Is(*matter) {
									return true
								}
							}
						}
					}
				}
				return false
			}

			if spawnItem.Placement.Air.Blocking != nil {
				if !checkMatter(y, x, z, spawnItem.Placement.Air.Blocking, false) {
					continue
				}
			}
			if spawnItem.Placement.Air.Matter != nil {
				if !checkMatter(y, x, z, spawnItem.Placement.Air.Matter, true) {
					continue
				}
			}
			if spawnItem.Placement.Surface.Blocking != nil {
				if !checkMatter(y-1, x, z, spawnItem.Placement.Surface.Blocking, false) {
					continue
				}
			}
			if spawnItem.Placement.Surface.Matter != nil {
				if !checkMatter(y-1, x, z, spawnItem.Placement.Surface.Matter, true) {
					continue
				}
			}

			s := spawn(y, x, z)
			if s == 2 {
				failed = true
				break
			} else if s == 1 {
			} else {
				break
			}
		}
		if failed {
			break
		}
	}
	return nil
}
