package world

import (
	"time"
)

// StatusI is the basic interface for Status access.
type StatusI interface {
	Originator() ObjectI
	SetOriginator(ObjectI)
	Target() ObjectI
	SetTarget(ObjectI)
	ShouldRemove() bool
	update(time.Duration)
}
