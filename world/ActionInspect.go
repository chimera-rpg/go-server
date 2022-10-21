package world

import "time"

// ActionInspect represents an action to inspect another object.
type ActionInspect struct {
	Action
	Target ID
}

// NewActionInspect returns an instantized version of ActionInspect
func NewActionInspect(target ID, cost time.Duration) *ActionInspect {
	return &ActionInspect{
		// FIXME: Figure out the actual base costs of inspection.
		Action: Action{
			channel:  cost / 4,
			recovery: cost - cost/4,
		},
		Target: target,
	}
}
