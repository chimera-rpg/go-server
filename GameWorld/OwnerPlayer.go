package GameWorld

type OwnerPlayer struct {
  ClientConnection ClientConnection
  target ObjectI
}

func (player OwnerPlayer) getTarget() ObjectI {
  return player.target
}
func (player OwnerPlayer) setTarget(object ObjectI) {
  player.target = object
}

func NewOwnerPlayer(cc ClientConnection) *OwnerPlayer {
  return &OwnerPlayer{
    ClientConnection: cc,
  }
}

func (player *OwnerPlayer) Update(delta int64) error {
  return nil
}


type ClientConnection interface {
}
