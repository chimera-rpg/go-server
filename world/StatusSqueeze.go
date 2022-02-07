package world

import (
	"math"
	"time"

	cdata "github.com/chimera-rpg/go-common/data"
)

// StatusSqueeze is the status for when an object is to enter the squeezing state.
type StatusSqueeze struct {
	Status
	Activate  bool
	X, Z      int
	Squeezing bool
	Remove    bool
}

var StatusSqueezeRef = &StatusSqueeze{}

func (s *StatusSqueeze) update(delta time.Duration) {
	s.Status.update(delta)
	if s.target == nil {
		s.shouldRemove = true
		return
	}
	// I guess this can only be toggled on/off by the owner. Perhaps should just be checked for in collision checks?
}

// StatusType returns cdata.CrouchingStatus
func (s *StatusSqueeze) StatusType() cdata.StatusType {
	return cdata.SqueezingStatus
}

// OnAdd calculates and stores our desired squeezing width and height delta from the target object's archetype.
func (s *StatusSqueeze) OnAdd() {
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
