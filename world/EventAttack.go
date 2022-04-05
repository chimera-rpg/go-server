package world

import (
	"fmt"
)

// EventAttack is emitted when an object is attacking another.
type EventAttacking struct {
	Target ObjectI
	// Damage
}

// String returns a string representing the attack.
func (e EventAttacking) String() string {
	return fmt.Sprintf("You attack %s", e.Target.Name())
}

// EventAttacked is emitted when an object is attacked.
type EventAttacked struct {
	Attacker ObjectI
	// Damage
	Prevented bool // Prevented flags the damage to not be applied, but still will notify the attacker of their damage.
}

// String returns a string representing the attack.
func (e EventAttacked) String() string {
	return fmt.Sprintf("You are attacked by %s", e.Attacker.Name())
}

// EventAttack is emitted when an object attacks another.
type EventAttack struct {
	Target ObjectI
	// Damage
}

// String returns a string representing the attack.
func (e EventAttack) String() string {
	return fmt.Sprintf("You attacked %s", e.Target.Name())
}
