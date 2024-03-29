package world

import (
	"math"
	"time"

	"github.com/chimera-rpg/go-server/data"
)

// StatusCrouch is the status for when an object is in the crouching state.
type StatusCrouch struct {
	Status
	Y         int
	Activate  bool
	Precrouch bool
	Crouching bool
	Remove    bool
}

var StatusCrouchRef = &StatusCrouch{}

func (s *StatusCrouch) update(delta time.Duration) {
	s.Status.update(delta)
	if s.target == nil {
		s.shouldRemove = true
		return
	}
}

// StatusType returns data.CrouchingStatus
func (s *StatusCrouch) StatusType() data.StatusType {
	return data.CrouchingStatus
}

// OnAdd calculates and stores our desired squeezing width and height delta from the target object's archetype.
func (s *StatusCrouch) OnAdd() {
	var h int
	a := s.target.GetArchetype()
	if a != nil {
		h = int(a.Height)

		s.Y = int(math.Max(float64(h)/2, 1))
		if s.Y == h {
			s.Y = 0
		}
	}
}
