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

// OwnerInspectCommand represents an inspection request for the given target.
type OwnerInspectCommand struct {
	Target ID
}

// OwnerEquipCommand represents an equip request for the given target.
type OwnerEquipCommand struct {
	Container ID   // The container to use. If none is specified, it is presumed the container is the character's default inventory.
	Y, X, Z   int  // The Y, X, Z location to look at for a container, if a container ID is specified.
	Target    ID   // The target item to equip.
	Equip     bool // Whether or not the target should be equipped or unequipped. If false, then it is assumed the item is in the player's equipment list and will be removed + placed into the player's container.
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
