package world

import "time"

// ActionAttack represents an action that is an attack.
type ActionAttack struct {
	Action
	Target  ID
	Y, X, Z int
	// Type int -- swing, pierce, etc.
}

// NewActionAttack returns an instantized version of ActionAttack
func NewActionAttack(y, x, z int, target ID, cost time.Duration) *ActionAttack {
	return &ActionAttack{
		Action: Action{
			channel:  cost / 4,
			recovery: cost - cost/4,
		},
		Y:      y,
		X:      x,
		Z:      z,
		Target: target,
	}
}
