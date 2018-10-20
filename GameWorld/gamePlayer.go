package GameWorld

type Player struct {
  ClientConnection ClientConnection
  target *GameObject
}

func (player Player) getTarget() *GameObject {
  return player.target
}
func (player Player) setTarget(object *GameObject) {
  player.target = object
}

func NewPlayer(cc ClientConnection) *Player {
  return &Player{
    ClientConnection: cc,
  }
}

func (player *Player) Update(delta int64) error {
  return nil
}


type ClientConnection interface {
}
