package world

import "time"

// ActionStatus is the action for changing status.
type ActionStatus struct {
	Action
	status StatusI
}

// NewActionStatus returns a populated ActionStatus.
func NewActionStatus(status StatusI, duration time.Duration) *ActionStatus {
	return &ActionStatus{
		Action: Action{
			channel: duration,
		},
		status: status,
	}
}
