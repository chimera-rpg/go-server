package world

import (
	cdata "github.com/chimera-rpg/go-common/data"
)

// StatusWizard is the status for when an object is flying.
type StatusWizard struct {
	Status
}

var StatusWizardRef = &StatusWizard{}

// StatusType returns cdata.WizardStatus
func (s *StatusWizard) StatusType() cdata.StatusType {
	return cdata.WizardStatus
}
