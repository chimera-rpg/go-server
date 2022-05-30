package data

import (
	"math/rand"
	"time"
)

// TimeRange just defines a range between two times.
type TimeRange struct {
	Min int `json:"Min" yaml:"Min,omitempty"`
	Max int `json:"Max" yaml:"Max,omitempty"`
}

// Random returns a duration between min and max.
func (t TimeRange) Random() time.Duration {
	if t.Min == t.Max {
		return time.Duration(t.Max) * time.Second
	}
	return time.Duration(rand.Intn(int(t.Max)-int(t.Min))+int(t.Min)) * time.Second
}

// ArchetypeTimer represents a built-in timer for an archetype. The result of these timers will be the triggering of an Event.
type ArchetypeTimer struct {
	Event  string    `json:"Event" yaml:"Event,omitempty"`
	Repeat int       `json:"Repeat" yaml:"Repeat,omitempty"`
	Wait   TimeRange `json:"Wait" yaml:"Wait,omitempty"`
}
