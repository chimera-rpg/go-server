package world

// OwnerPlayer represents a player character through a network
// connection and the associated player object.
type OwnerPlayer struct {
	ClientConnection ClientConnection
	target           ObjectI
}

func (player OwnerPlayer) getTarget() ObjectI {
	return player.target
}
func (player OwnerPlayer) setTarget(object ObjectI) {
	player.target = object
}

// NewOwnerPlayer creates a Player from a given client connection.
func NewOwnerPlayer(cc ClientConnection) *OwnerPlayer {
	return &OwnerPlayer{
		ClientConnection: cc,
	}
}

// Update does something.?
func (player *OwnerPlayer) Update(delta int64) error {
	return nil
}

// ClientConnection is an interface? We should probably use common/net or something.
type ClientConnection interface {
}
