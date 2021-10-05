package world

import (
	"time"

	cdata "github.com/chimera-rpg/go-common/data"
)

// StatusFalling is the status for when an object is falling.
type StatusFalling struct {
	Status
	aggregate    time.Duration
	fallDistance int
}

func (s *StatusFalling) update(delta time.Duration) {
	const fallRate = 4717 * time.Microsecond * 10 // 53 meters/second or 212 units/second or 1 unit/4717 microseconds. We multiply this by 10 so you don't just reach terminal instantly (and so it feels nicer).

	s.Status.update(delta)
	if s.target == nil {
		s.shouldRemove = true
		return
	}

	doTilesBlock := func(targetTiles []*Tile) bool {
		isBlocked := false
		matter := s.target.GetArchetype().Matter
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
	s.aggregate += delta
	for s.aggregate >= fallRate {
		s.aggregate -= fallRate
		m := s.target.GetTile().gameMap
		if m != nil {
			_, fallingTiles, err := m.GetObjectPartTiles(s.target, -1, 0, 0, false)

			if doTilesBlock(fallingTiles) && err == nil {
				if s.fallDistance >= 4 {
					s.target.ResolveEvent(EventFell{
						distance: int(s.elapsed / fallRate),
					})
				}
				s.shouldRemove = true
				return
			}
			if _, err := m.MoveObject(s.target, -1, 0, 0, true); err != nil {
				// Remove status if we had an error while moving.
				s.shouldRemove = true
				return
			}
			s.fallDistance++
			if s.fallDistance == 4 {
				s.target.ResolveEvent(EventFall{})
			}
		}
	}
}

// StatusType returns cdata.Falling
func (s *StatusFalling) StatusType() cdata.StatusType {
	return cdata.FallingStatus
}
