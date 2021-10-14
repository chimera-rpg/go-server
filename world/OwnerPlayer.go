package world

import (
	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-common/network"
	"github.com/chimera-rpg/go-server/data"

	"time"

	log "github.com/sirupsen/logrus"
)

type clientConnectionI interface {
	GetUser() *data.User
	SetOwner(p *OwnerPlayer)
	GetOwner() *OwnerPlayer
	Send(network.Command) error
	GetID() int
}

// OwnerPlayer represents a player character through a network
// connection and the associated player object.
type OwnerPlayer struct {
	Owner
	commandChannel                   chan OwnerCommand
	ClientConnection                 clientConnectionI
	target                           *ObjectCharacter
	currentMap                       *Map
	mapUpdateTime                    uint8
	viewWidth, viewHeight, viewDepth int
	view                             [][][]TileView
	knownIDs                         map[ID]struct{}
	attitudes                        map[ID]data.Attitude
	lastKnownStamina                 time.Duration
}

// GetTarget returns the player's target object.
func (player *OwnerPlayer) GetTarget() ObjectI {
	return player.target
}

// SetTarget sets the given object as the target of the player.
func (player *OwnerPlayer) SetTarget(object ObjectI) {
	if objectpc, ok := object.(*ObjectCharacter); ok {
		player.target = objectpc
	} else {
		log.Printf("Attempted to set OwnerPlayer to non-ObjectCharacter...\n")
	}
	object.SetOwner(player)
}

// GetCommandChannel gets the command channel for the player.
func (player *OwnerPlayer) GetCommandChannel() chan OwnerCommand {
	return player.commandChannel
}

// GetMap gets the currentMap of the owner.
func (player *OwnerPlayer) GetMap() *Map {
	return player.currentMap
}

// SetMap sets the currentMap of the owner.
func (player *OwnerPlayer) SetMap(m *Map) {
	player.currentMap = m
	// Might as well let the client know what's up.
	if m != nil {
		player.ClientConnection.Send(network.CommandMap{
			Name:   m.name,
			MapID:  m.mapID,
			Height: m.height,
			Width:  m.width,
			Depth:  m.depth,
		})
	}
	// Create a fresh view corresponding to our new map.
	player.CreateView()
}

// NewOwnerPlayer creates a Player from a given client connection.
func NewOwnerPlayer(cc clientConnectionI) *OwnerPlayer {
	return &OwnerPlayer{
		commandChannel:   make(chan OwnerCommand),
		ClientConnection: cc,
		knownIDs:         make(map[ID]struct{}),
		viewWidth:        32,
		viewHeight:       16,
		viewDepth:        32,
	}
}

// CreateView creates the initial view of the player.
func (player *OwnerPlayer) CreateView() {
	gmap := player.GetMap()
	if gmap == nil {
		player.view = make([][][]TileView, 0)
		return
	}
	player.view = make([][][]TileView, gmap.height)
	for y := 0; y < gmap.height; y++ {
		player.view[y] = make([][]TileView, gmap.width)
		for x := 0; x < gmap.width; x++ {
			player.view[y][x] = make([]TileView, gmap.depth)
		}
	}
}

// CheckView checks the view around the player and calls any associated network functions to update the client.
func (player *OwnerPlayer) CheckView() {
	// Map has changed in some way, so let's check our viewable tiles for updates.
	player.checkVisibleTiles()
}

// SetViewSize sets the viewport limits of the player.
func (player *OwnerPlayer) SetViewSize(h, w, d int) {
	player.viewHeight = h
	player.viewWidth = w
	player.viewDepth = d
}

// GetViewSize returns the view port size that is used to send map updates to the player.
func (player *OwnerPlayer) GetViewSize() (h, w, d int) {
	// TODO: Probably conditionally replace with target object's vision.
	return player.viewHeight, player.viewWidth, player.viewDepth
}

// checkVisibleTiles gets the initial view of the player. This sends tile information equal to how far the owner's PC can see.
func (player *OwnerPlayer) checkVisibleTiles() error {
	gmap := player.GetMap()
	// Get owner's viewport.
	vh, vw, vd := player.GetViewSize()
	vhh := vh / 2
	vwh := vw / 2
	vdh := vd / 2
	// Get tile where owner is, then send from negative half owner object's viewport to positive half in y, x, and z.
	if tile := player.target.GetTile(); tile != nil {
		var sy, sx, sz, ey, ex, ez int
		if sy = tile.y - vhh; sy < 0 {
			sy = 0
		}
		if sx = tile.x - vwh; sx < 0 {
			sx = 0
		}
		if sz = tile.z - vdh; sz < 0 {
			sz = 0
		}
		if ey = tile.y + vhh; ey > len(player.view) {
			ey = len(player.view)
		}
		if ex = tile.x + vwh; ex > len(player.view[0]) {
			ex = len(player.view[0])
		}
		if ez = tile.z + vdh; ez > len(player.view[0][0]) {
			ez = len(player.view[0][0])
		}

		for yi := sy; yi < ey; yi++ {
			for xi := sx; xi < ex; xi++ {
				for zi := sz; zi < ez; zi++ {
					// TODO: Actually calculate which tiles are visible for the owner.
					mapTile := gmap.GetTile(yi, xi, zi)
					if mapTile.modTime == player.view[yi][xi][zi].modTime {
						continue
					}
					player.view[yi][xi][zi].modTime = mapTile.modTime
					// NOTE: We could maintain a list of known objects on a tile in the player's tile view and send the difference instead. For large stacks of infrequently changing tiles, this would be more bandwidth efficient, though at the expense of server-side RAM and CPU time.
					// Filter out things we don't want to send to the client.
					filteredMapObjects := make([]ObjectI, 0)
					for _, o := range mapTile.GetObjects() {
						if o.getType() != cdata.ArchetypeAudio {
							filteredMapObjects = append(filteredMapObjects, o)
						}
					}
					tileObjectIDs := make([]ID, len(filteredMapObjects))
					// Send any objects unknown to the client (and collect their IDs).
					for i, o := range filteredMapObjects {
						oID := o.GetID()
						if _, isObjectKnown := player.knownIDs[oID]; !isObjectKnown {
							// Let the client know of the object(s). NOTE: We could send a collection of object creation commands so as to reduce TCP overhead for bulk updates.
							oArch := o.GetArchetype()
							if oArch != nil {
								player.ClientConnection.Send(network.CommandObject{
									ObjectID: o.GetID(),
									Payload: network.CommandObjectPayloadCreate{
										TypeID:      o.getType().AsUint8(),
										AnimationID: oArch.AnimID,
										FaceID:      oArch.FaceID,
										Height:      oArch.Height,
										Width:       oArch.Width,
										Depth:       oArch.Depth,
									},
								})
							} else {
								player.ClientConnection.Send(network.CommandObject{
									ObjectID: o.GetID(),
									Payload:  network.CommandObjectPayloadCreate{},
								})
							}
							player.knownIDs[oID] = struct{}{}
						}
						tileObjectIDs[i] = oID
					}
					// Update the client's perception of the given tile.
					player.ClientConnection.Send(network.CommandTile{
						Y:         uint32(yi),
						X:         uint32(xi),
						Z:         uint32(zi),
						ObjectIDs: tileObjectIDs,
					})
				}
			}
		}
	}

	return nil
}

// Update does something.?
func (player *OwnerPlayer) Update(delta time.Duration) error {
	done := false
	for !done {
		select {
		case ocmd, _ := <-player.commandChannel:
			switch c := ocmd.(type) {
			case OwnerRepeatCommand:
				if c.Cancel {
					player.repeatCommand = nil
				} else {
					player.repeatCommand = c
				}
			case OwnerClearCommand:
				player.ClearCommands()
			default:
				player.PushCommand(c)
			}
		default:
			done = true
		}
	}

	// TODO: Throttle sending updates for stamina and others.
	if t := player.GetTarget(); t != nil {
		/*if t.Stamina() != player.lastKnownStamina {
			player.lastKnownStamina = t.Stamina()
			player.ClientConnection.Send(network.CommandStamina{
				Stamina:    player.lastKnownStamina,
				MaxStamina: t.MaxStamina(),
			})
		}*/
	}

	return nil
}

// OnMapUpdate is called when the map is updated and the player should update its view and/or react.
func (player *OwnerPlayer) OnMapUpdate(delta time.Duration) error {
	if player.mapUpdateTime == player.currentMap.updateTime {
		return nil
	}
	player.CheckView()

	// Make sure we're in sync.
	player.mapUpdateTime = player.currentMap.updateTime

	return nil
}

// OnObjectDelete is called when an object on the map is deleted. If the player knows about it, then an object delete command is sent to the client.
func (player *OwnerPlayer) OnObjectDelete(oID ID) error {
	if _, isObjectKnown := player.knownIDs[oID]; isObjectKnown {
		player.ClientConnection.Send(network.CommandObject{
			ObjectID: oID,
			Payload:  network.CommandObjectPayloadDelete{},
		})
		delete(player.knownIDs, oID)
	}

	return nil
}

// GetAttitude returns the attitude the owner has the a given object. If no attitude exists, one is calculated based upon the target's attitude (if it has one).
func (player *OwnerPlayer) GetAttitude(oID ID) data.Attitude {
	if attitude, ok := player.attitudes[oID]; ok {
		return attitude
	}
	target := player.GetMap().world.GetObject(oID)
	if target == nil {
		delete(player.attitudes, oID)
	} else {
		// TODO: We should probably check if the target knows us and use their attitude. If not, we should calculate from our target object archetype's default attitude towards: Genera, Species, Legacy, and Faction.
		if otherOwner := target.GetOwner(); otherOwner != nil {
			return otherOwner.GetAttitude(player.target.id)
		}
	}

	return data.NoAttitude
}

// SendMessage sends a message to the character.
func (player *OwnerPlayer) SendMessage(s string) {
	player.ClientConnection.Send(network.CommandMessage{
		Type:         network.TargetMessage,
		FromObjectID: player.target.id,
		Body:         s,
	})
}

// SendStatus sends the status to the owner, providing it has a StatusType.
func (player *OwnerPlayer) SendStatus(s StatusI, active bool) {
	if s.StatusType() != 0 {
		player.ClientConnection.Send(network.CommandStatus{
			Type:   s.StatusType(),
			Active: active,
		})
	}
}

// SendSound sends the given sound to the player.
func (player *OwnerPlayer) SendSound(audioID ID, soundID ID, objectID ID, y, x, z int, volume float32) {
	player.ClientConnection.Send(network.CommandNoise{
		Type:     network.GenericNoise,
		AudioID:  audioID,
		SoundID:  soundID,
		ObjectID: objectID,
		Y:        uint32(y),
		X:        uint32(x),
		Z:        uint32(z),
		Volume:   volume,
	})
}

// SendMusic sends music.
func (player *OwnerPlayer) SendMusic(audioID, soundID ID, soundIndex int8, objectID ID, y, x, z int, volume float32, loopCount int8) {
	player.ClientConnection.Send(network.CommandMusic{
		Type:     network.MapNoise,
		AudioID:  audioID,
		SoundID:  soundID,
		ObjectID: objectID,
		Y:        uint32(y),
		X:        uint32(x),
		Z:        uint32(z),
		Volume:   volume,
		Loop:     loopCount,
	})
}

// StopMusic stops music playing from a source.
func (player *OwnerPlayer) StopMusic(objectID ID) {
	player.ClientConnection.Send(network.CommandMusic{
		Type:     network.MapNoise,
		ObjectID: objectID,
		Stop:     true,
	})
}
