package world

import (
	"time"
)

// StatusSqueezing is the status for when an object is in the squeezing state.
type StatusSqueezing struct {
	Status
}

func (s *StatusSqueezing) update(delta time.Duration) {
	s.Status.update(delta)
	if s.target == nil {
		s.shouldRemove = true
		return
	}
	// I guess this can only be toggled on/off by the owner. Perhaps should just be checked for in collision checks?
}
