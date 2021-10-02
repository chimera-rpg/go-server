package world

import (
	"time"
)

// StatusUnsqueeze is the status for when an object is to leave the squeezing state.
type StatusUnsqueeze struct {
	Status
}

func (s *StatusUnsqueeze) update(delta time.Duration) {
	s.Status.update(delta)
	if s.target == nil {
		s.shouldRemove = true
		return
	}
}
