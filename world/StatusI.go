package world

import (
	"time"

	"github.com/chimera-rpg/go-server/data"
)

// StatusI is the basic interface for Status access.
type StatusI interface {
	Originator() ObjectI
	SetOriginator(ObjectI)
	Target() ObjectI
	SetTarget(ObjectI)
	ShouldRemove() bool
	update(time.Duration)
	StatusType() data.StatusType
	OnAdd()
	OnRemove()
}
