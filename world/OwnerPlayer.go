package world

type clientConnectionI interface {
	SetOwner(p *OwnerPlayer)
	GetOwner() *OwnerPlayer
}

// OwnerPlayer represents a player character through a network
// connection and the associated player object.
type OwnerPlayer struct {
	ClientConnection clientConnectionI
	target           ObjectI
}

func (player OwnerPlayer) getTarget() ObjectI {
	return player.target
}
func (player OwnerPlayer) setTarget(object ObjectI) {
	player.target = object
	object.setOwner(player)
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
