package world

type clientConnectionI interface {
	SetOwner(p OwnerI)
	GetOwner() OwnerI
}

// OwnerPlayer represents a player character through a network
// connection and the associated player object.
type OwnerPlayer struct {
	ClientConnection clientConnectionI
	target           ObjectI
}

// GetTarget returns the player's target object.
func (player *OwnerPlayer) GetTarget() ObjectI {
	return player.target
}

// SetTarget sets the given object as the target of the player.
func (player *OwnerPlayer) SetTarget(object ObjectI) {
	player.target = object
	object.SetOwner(player)
}

// NewOwnerPlayer creates a Player from a given client connection.
func NewOwnerPlayer(cc clientConnectionI) *OwnerPlayer {
	return &OwnerPlayer{
		ClientConnection: cc,
	}
}

// Update does something.?
func (player *OwnerPlayer) Update(delta int64) error {
	// I guess here is where we'd have some sort of "handleCommandQueue" functionality.
	return nil
}
