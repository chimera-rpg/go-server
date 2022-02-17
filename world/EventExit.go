package world

// EventExit is emitted when a teleporter object is used. If prevent is set to true, then the exit action is prevented.
type EventExit struct {
	Target  ObjectI
	Prevent bool
	Message string
}
