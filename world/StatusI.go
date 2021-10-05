package world

import (
	"time"

	cdata "github.com/chimera-rpg/go-common/data"
)

// StatusI is the basic interface for Status access.
type StatusI interface {
	Originator() ObjectI
	SetOriginator(ObjectI)
	Target() ObjectI
	SetTarget(ObjectI)
	ShouldRemove() bool
	update(time.Duration)
	StatusType() cdata.StatusType
	OnAdd()
	OnRemove()
}
