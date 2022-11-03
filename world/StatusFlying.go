package world

import (
	"github.com/chimera-rpg/go-server/data"
)

// StatusFlying is the status for when an object is flying.
type StatusFlying struct {
	Status
}

var StatusFlyingRef = &StatusFlying{}

// StatusType returns data.FlyingStatus
func (s *StatusFlying) StatusType() data.StatusType {
	return data.FlyingStatus
}
