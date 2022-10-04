package world

import "time"

// Updates represents the world-wide updates for a given tick.
type Updates struct {
	Delta   time.Duration
	Updates []Update
}

// Update is an interface for world-wide updates.
type Update interface {
	// ???
}
