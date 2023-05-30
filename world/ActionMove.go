package world

import "time"

// ActionMove represents an action that be a move.,
type ActionMove struct {
	Action
	x, y, z int
	running bool
}

// NewActionMove returns an instantized version of ActionMove
func NewActionMove(y, x, z int, cost time.Duration, running bool) *ActionMove {
	return &ActionMove{
		Action: Action{
			channel:  cost / 4,
			recovery: cost - cost/4,
		},
		y:       y,
		x:       x,
		z:       z,
		running: running,
	}
}

func (m *Map) HandleActionMove(a *ActionMove) error {
	_, err := a.object.GetTile().GetMap().MoveObject(a.object, a.y, a.x, a.z, false)
	return err
}
