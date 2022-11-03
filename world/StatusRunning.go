package world

import (
	"github.com/chimera-rpg/go-server/data"
)

// StatusRunning is the status for when an object is running.
type StatusRunning struct {
	Status
}

var StatusRunningRef = &StatusRunning{}

// StatusType returns data.CrouchingStatus
func (s *StatusRunning) StatusType() data.StatusType {
	return data.RunningStatus
}
