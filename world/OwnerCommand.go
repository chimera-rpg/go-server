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

// OwnerGrabCommand reqpresents a command to pick up a target.
type OwnerGrabCommand struct {
	FromContainer ID // The container to grab from. If none is specified it is presumed to be an item in reach. If specified, first the character's containers are checked, then the object in the world if in reach.
	ToContainer   ID // The target container to place the item into. First the character's containers are checked, then the object in the world if in reach.
	Target        ID // The target item to grab. If FromContainer is 0, then it is searched for if in reach.
}

// OwnerDropCommand represents a command to place a target onto the ground.
type OwnerDropCommand struct {
	FromContainer ID  // The container to drop from. If none is specified, it is presumed to be the character's base inventory. If it is non-zero, first the character's containers are checked for a match, then the object in the world (and if it is within reach).
	Y, X, Z       int // The target position to drop to.
	Target        ID  // The target item to drop.
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
