package world

// OwnerCommand does stuff.
type OwnerCommand interface {
}

// OwnerMoveCommand represents an owner command to move an object towards a given direction.
type OwnerMoveCommand struct {
	Y, X, Z   int
	Running   bool // Whether or not it is a running move.
	Canceling bool // Whether or not it is a canceling move.
}

// OwnerStatusCommand represents an owner command to set the status for its target object.
type OwnerStatusCommand struct {
	Status StatusI
}

// OwnerRepeatCommand embeds repeating commands.
type OwnerRepeatCommand struct {
	Command OwnerCommand
	Cancel  bool
}

// OwnerAttackCommand
type OwnerAttackCommand struct {
	Direction int
	Y, X, Z   int
	Target    ID
}

// OwnerClearCommand clears the owner's commands.
type OwnerClearCommand struct{}

// OwnerWizardCommand toggles wizard mode.
type OwnerWizardCommand struct{}

// OwnerExtCommand handles extended commands.
type OwnerExtCommand struct {
	Command string
	Args    []string
}
