package world

import (
	"fmt"
)

type EventFell struct {
	distance int
}

func (e EventFell) String() string {
	return fmt.Sprintf("You fell %d meters", (e.distance)/4.0)
}

type EventFall struct {
}
