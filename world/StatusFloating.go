package world

import (
	cdata "github.com/chimera-rpg/go-common/data"
)

// StatusFloating is the status for when an object is floating.
type StatusFloating struct {
	Status
}

// StatusType returns cdata.FloatingStatus
func (s *StatusFloating) StatusType() cdata.StatusType {
	return cdata.FloatingStatus
}
