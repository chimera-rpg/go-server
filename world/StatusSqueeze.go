package world

import (
	"time"
)

// StatusSqueeze is the status for when an object is to enter the squeezing state.
type StatusSqueeze struct {
	Status
}

func (s *StatusSqueeze) update(delta time.Duration) {
	s.Status.update(delta)
	if s.target == nil {
		s.shouldRemove = true
		return
	}
	// I guess this can only be toggled on/off by the owner. Perhaps should just be checked for in collision checks?
}
