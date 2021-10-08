package world

import "time"

// Action is our base action type.
type Action struct {
	channel, recovery time.Duration
	channeled         bool
}

// ChannelTime returns the channel time of the action.
func (a *Action) ChannelTime() time.Duration {
	return a.channel
}

// RecoveryTime returns the rest time of the action.
func (a *Action) RecoveryTime() time.Duration {
	return a.recovery
}

// Channel sets the channeled state of the action.
func (a *Action) Channel(bool) {
	a.channeled = true
}

// Channeled returns the channeled status of the action.
func (a *Action) Channeled() bool {
	return a.channeled
}
