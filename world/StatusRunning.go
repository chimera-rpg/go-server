package world

import (
	cdata "github.com/chimera-rpg/go-common/data"
)

// StatusRunning is the status for when an object is running.
type StatusRunning struct {
	Status
}

var StatusRunningRef = &StatusRunning{}

// StatusType returns cdata.CrouchingStatus
func (s *StatusRunning) StatusType() cdata.StatusType {
	return cdata.RunningStatus
}
