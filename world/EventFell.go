package world

import (
	"fmt"

	"github.com/chimera-rpg/go-server/data"
)

// EventFell is emitted when an object has fallen a given distance.
type EventFell struct {
	distance int
	matter   data.MatterType
}

// String returns a string representing how far the target fell in the second person.
func (e EventFell) String() string {
	return fmt.Sprintf("You fell %d meters", (e.distance)/4.0)
}

// EventFall is emitted when an object is falling.
type EventFall struct {
}
