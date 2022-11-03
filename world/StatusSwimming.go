package world

import (
	data "github.com/chimera-rpg/go-server/data"
)

// StatusSwimming is the status for when an object is swimming.
type StatusSwimming struct {
	Status
}

var StatusSwimmingRef = &StatusSwimming{}

// StatusType returns data.SwimmingStatus
func (s *StatusSwimming) StatusType() data.StatusType {
	return data.SwimmingStatus
}
