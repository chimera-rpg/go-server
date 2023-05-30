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
		targetObject := m.world.GetObject(a.Target)
		if targetObject == nil {
			return errors.New("object does not exist")
		}
		t := targetObject.GetTile()
		if container := targetObject.GetContainer(); container != nil {
			// Always consider it near if its in a container.
			near = true
			// Always send mundane info.
			mundaneInfo := targetObject.GetMundaneInfo(near)
			mundaneInfo.Near = near
			infos = append(infos, mundaneInfo)
			// If it's a contained item, check if
			if container == a.object {
				// It's the player's base inventory
			} else {
				if targetObject, _ := o.FeatureInventory.GetObjectByID(a.Target); targetObject != nil {
					// It's in one of the player's held containers.
				} else {
					// It might be in the world!
					if container.GetTile().gameMap != o.tile.gameMap {
						return errors.New("container not in same map")
					}
					if !o.InReachRange(container.GetTile().Y, container.GetTile().X, container.GetTile().Z) {
						return errors.New("container is out of range")
					}
				}
			}
		} else if t.GetMap() == o.tile.gameMap {
			if o.InReachRange(t.Y, t.X, t.Z) {
				// TODO: Do a line of sight check from the character's intersection cube.
				near = true
				// Send detailed info?
			}
			// Always get the mundane info.
			mundaneInfo := targetObject.GetMundaneInfo(near)
			mundaneInfo.Near = near
			infos = append(infos, mundaneInfo)
		}
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
		return nil
	}
	return errors.New("object is not a character")
}
