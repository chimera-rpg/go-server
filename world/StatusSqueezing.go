package world

import (
	"math"
	"time"

	cdata "github.com/chimera-rpg/go-common/data"
)

// StatusSqueezing is the status for when an object is in the squeezing state.
type StatusSqueezing struct {
	Status
	X, Z int
}

func (s *StatusSqueezing) update(delta time.Duration) {
	s.Status.update(delta)
	if s.target == nil {
		s.shouldRemove = true
		return
	}
	// I guess this can only be toggled on/off by the owner. Perhaps should just be checked for in collision checks?
}

// StatusType returns cdata.Squeezing
func (s *StatusSqueezing) StatusType() cdata.StatusType {
	return cdata.SqueezingStatus
}

// OnAdd calculates and stores our desired squeezing width and height delta from the target object's archetype.
func (s *StatusSqueezing) OnAdd() {
	var w, d int
	a := s.target.GetArchetype()
	if a != nil {
		w = int(a.Width)
		d = int(a.Depth)

		s.X = int(math.Max(float64(w)/3, 1))
		if s.X == w {
			s.X = 0
		}
		s.Z = int(math.Max(float64(d)/3, 1))
		if s.Z == d {
			s.Z = 0
		}
	}
}
