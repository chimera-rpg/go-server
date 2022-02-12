package world

import "time"

// Timer is used by objects to manage archetype timers.
type Timer struct {
	elapsed     time.Duration
	target      time.Duration
	event       string
	repeat      int
	repeatCount int
}
