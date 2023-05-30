package world

import (
	"errors"
	"time"

	"github.com/chimera-rpg/go-server/data"
	"github.com/chimera-rpg/go-server/network"
)

// ActionInspect represents an action to inspect another object.
type ActionInspect struct {
	Action
	Target ID
}

// NewActionInspect returns an instantized version of ActionInspect
func NewActionInspect(target ID, cost time.Duration) *ActionInspect {
	return &ActionInspect{
		// FIXME: Figure out the actual base costs of inspection.
		Action: Action{
			channel:  cost / 4,
			recovery: cost - cost/4,
		},
		Target: target,
	}
}

func (m *Map) HandleActionInspect(a *ActionInspect) error {
	if o, ok := a.object.(*ObjectCharacter); ok {
		var infos []data.ObjectInfo
		var near bool
		// TODO: Kick any owners who keep sending bad commands.
		if targetObject := m.world.GetObject(a.Target); targetObject != nil {
			t := targetObject.GetTile()
			if t.GetMap() == o.tile.gameMap {
				if o.InReachRange(t.Y, t.X, t.Z) {
					// TODO: Do a line of sight check from the character's intersection cube.
					near = true
					// Send detailed info?
				}
				// Always get the mundane info.
				mundaneInfo := targetObject.GetMundaneInfo(near)
				mundaneInfo.Near = near
				infos = append(infos, mundaneInfo)

				// Send any information if we have it.
				if len(infos) > 0 {
					// Only send to non-AI players.
					if p, ok := o.GetOwner().(*OwnerPlayer); ok {
						p.ClientConnection.Send(network.CommandObject{
							ObjectID: a.Target,
							Payload: network.CommandObjectPayloadInfo{
								Info: infos,
							},
						})
					}
				}
			}
		}
	}
	return errors.New("object is not a character")
}
