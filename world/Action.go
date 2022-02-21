package world

import "time"

// Action is our base action type.
type Action struct {
	object            ObjectI
	channel, recovery time.Duration
	channeled         bool
	ready             bool
}

// Object returns the action's associated object.
func (a *Action) Object() ObjectI {
	return a.object
}

// SetObject sets the action's object.
func (a *Action) SetObject(o ObjectI) {
	a.object = o
}

// Ready returns the action's ready state. If this is true, then the action will be processed by the owning map at the end of the update iteration.
func (a *Action) Ready() bool {
	return a.ready
}

// SetReadty sets the ready property to the passed value.
func (a *Action) SetReady(b bool) {
	a.ready = b
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
