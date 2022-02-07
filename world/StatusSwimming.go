package world

import (
	cdata "github.com/chimera-rpg/go-common/data"
)

// StatusSwimming is the status for when an object is swimming.
type StatusSwimming struct {
	Status
}

// StatusType returns cdata.SwimmingStatus
func (s *StatusSwimming) StatusType() cdata.StatusType {
	return cdata.SwimmingStatus
}
