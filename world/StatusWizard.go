package world

import (
	"github.com/chimera-rpg/go-server/data"
)

// StatusWizard is the status for when an object is flying.
type StatusWizard struct {
	Status
}

var StatusWizardRef = &StatusWizard{}

// StatusType returns data.WizardStatus
func (s *StatusWizard) StatusType() data.StatusType {
	return data.WizardStatus
}
