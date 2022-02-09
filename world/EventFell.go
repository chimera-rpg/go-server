package world

import (
	"fmt"

	cdata "github.com/chimera-rpg/go-common/data"
)

// EventFell is emitted when an object has fallen a given distance.
type EventFell struct {
	distance int
	matter   cdata.MatterType
}

// String returns a string representing how far the target fell in the second person.
func (e EventFell) String() string {
	return fmt.Sprintf("You fell %d meters", (e.distance)/4.0)
}

// EventFall is emitted when an object is falling.
type EventFall struct {
}
