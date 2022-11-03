package world

import (
	data "github.com/chimera-rpg/go-server/data"
)

// StatusFloating is the status for when an object is floating.
type StatusFloating struct {
	Status
}

var StatusFloatingRef = &StatusFloating{}

// StatusType returns data.FloatingStatus
func (s *StatusFloating) StatusType() data.StatusType {
	return data.FloatingStatus
}
