package world

import (
	"time"

	data "github.com/chimera-rpg/go-server/data"
)

// Status represents a pending status effect for an object. These tick over time.
type Status struct {
	elapsed      time.Duration
	originator   ObjectI
	target       ObjectI
	shouldRemove bool
}

// SetOriginator sets the originator of the status.
func (s *Status) SetOriginator(o ObjectI) {
	s.originator = o
}

// Originator returns the originator of the status.
func (s *Status) Originator() ObjectI {
	return s.originator
}

// SetTarget sets the target of the status.
func (s *Status) SetTarget(o ObjectI) {
	s.target = o
}

// Target returns the target of the status.
func (s *Status) Target() ObjectI {
	return s.target
}

func (s *Status) update(delta time.Duration) {
	s.elapsed += delta
}

// ShouldRemove returns if the status should be removed.
func (s *Status) ShouldRemove() bool {
	return s.shouldRemove
}

// StatusType returns the StatusType of the status.
func (s *Status) StatusType() data.StatusType {
	return 0
}

// OnAdd is called when the status is added. Target is guaranteed to be set.
func (s *Status) OnAdd() {
	return
}

// OnRemove is called when the status is removed. Target is guaranteed to be set.
func (s *Status) OnRemove() {
	return
}
