package world

import (
	cdata "github.com/chimera-rpg/go-common/data"
)

// StatusFlying is the status for when an object is flying.
type StatusFlying struct {
	Status
}

var StatusFlyingRef = &StatusFlying{}

// StatusType returns cdata.FlyingStatus
func (s *StatusFlying) StatusType() cdata.StatusType {
	return cdata.FlyingStatus
}
