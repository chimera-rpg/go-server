package world

import (
	"time"
)

type Status struct {
	elapsed      time.Duration
	originator   ObjectI
	target       ObjectI
	shouldRemove bool
}

func (s *Status) SetOriginator(o ObjectI) {
	s.originator = o
}

func (s *Status) Originator() ObjectI {
	return s.originator
}

func (s *Status) SetTarget(o ObjectI) {
	s.target = o
}

func (s *Status) Target() ObjectI {
	return s.target
}

func (s *Status) update(delta time.Duration) {
	s.elapsed += delta
}

func (s *Status) ShouldRemove() bool {
	return s.shouldRemove
}
