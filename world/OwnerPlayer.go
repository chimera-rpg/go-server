package world

import (
	"strings"

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
	mapUpdateTime                    uint8
	viewWidth, viewHeight, viewDepth int
	view                             [][][]TileView
	knownIDs                         map[ID]struct{}
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
	// Reset player's known IDs... TODO: Probably manage IDs on the client.
	player.knownIDs = make(map[uint32]struct{})
	// Create a fresh view corresponding to our new map.
	player.CreateView()
}

// NewOwnerPlayer creates a Player from a given client connection.
func NewOwnerPlayer(cc clientConnectionI) *OwnerPlayer {
	return &OwnerPlayer{
		Owner: Owner{
			attitudes: make(map[uint32]data.Attitude),
		},
		commandChannel:   make(chan OwnerCommand),
		ClientConnection: cc,
		knownIDs:         make(map[ID]struct{}),
		viewWidth:        48,
		viewHeight:       32,
		viewDepth:        48,
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
	//player.checkVisibleTiles()
	player.checkVisionRing()
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

// getVisionRing returns an array of coordinates matching the outermost edge of the player's vision.
func (player *OwnerPlayer) getVisionRing() (c [][3]int) {
	tile := player.GetTarget().GetTile()
	m := tile.GetMap()
	vh, vw, vd := player.GetViewSize()
	vhh := vh / 2
	vwh := vw / 2
	vdh := vd / 2
	y1 := tile.Y + int(player.GetTarget().GetArchetype().Height)
	// TODO: Use target object's statistics for vision range.
	add := func(y, x, z int) {
		if y < 0 || x < 0 || z < 0 || y >= m.height || x >= m.width || z >= m.depth {
			return
		}
		c = append(c, [3]int{y, x, z})
	}
	// bottom & top
	for x := tile.X - vwh; x < tile.X+vwh; x++ {
		for z := tile.Z - vdh; z < tile.Z+vdh; z++ {
			add(y1-vhh, x, z)
			add(y1+vhh, x, z)
		}
	}
	// left & right
	for y := tile.Y - vhh; y < tile.Y+vhh; y++ {
		for z := tile.Z - vdh; z < tile.Z+vdh; z++ {
			add(y, tile.X-vwh, z)
			add(y, tile.X+vwh, z)
		}
	}
	// back & front
	for y := tile.Y - vhh; y < tile.Y+vhh; y++ {
		for x := tile.X - vwh; x < tile.X+vwh; x++ {
			add(y, x, tile.Z-vdh)
			add(y, x, tile.Z+vdh)
		}
	}
	return
}

func (player *OwnerPlayer) getVisionCube() (c [][3]int) {
	tile := player.GetTarget().GetTile()
	m := tile.GetMap()
	vh, vw, vd := player.GetViewSize()
	vhh := vh / 2
	vwh := vw / 2
	vdh := vd / 2
	y1 := tile.Y + int(player.GetTarget().GetArchetype().Height)

	ymin := y1 - vhh
	if ymin < 0 {
		ymin = 0
	}
	ymax := y1 + vhh
	if ymax > m.height {
		ymax = m.height - 1
	}

	xmin := tile.X - vwh
	if xmin < 0 {
		xmin = 0
	}
	xmax := tile.X + vwh
	if xmax > m.width {
		xmax = m.width - 1
	}

	zmin := tile.Z - vdh
	if zmin < 0 {
		zmin = 0
	}
	zmax := tile.Z + vdh
	if zmax > m.depth {
		zmax = m.depth - 1
	}

	for y := ymin; y < ymax; y += 2 {
		for x := xmin; x < xmax; x++ {
			for z := zmin; z < zmax; z++ {
				c = append(c, [3]int{y, x, z})
			}
		}
	}
	return
}

func (player *OwnerPlayer) checkVisionRing() error {
	gmap := player.GetMap()
	tile := player.GetTarget().GetTile()
	//coords := player.getVisionRing()
	coords := player.getVisionCube()

	// Ensure our own tile is updated.
	player.sendTile(tile)

	a := player.GetTarget().GetArchetype()

	// TODO: We should also shoot rays from the target's feet a short distance to ensure close objects are visible.
	// Amanatides & Woo
	y1 := float64(tile.Y + int(a.Height))
	if y1 >= float64(gmap.height) {
		y1 = float64(gmap.height - 1)
	}
	x1 := float64(tile.X) + float64(a.Width)/2
	z1 := float64(tile.Z) + float64(a.Depth)/2
	for _, c := range coords {
		for _, tile := range gmap.ShootRay(y1, x1, z1, float64(c[0]), float64(c[1]), float64(c[2]), func(t *Tile) bool {
			return !t.opaque
		}) {
			player.sendTile(tile)
		}
	}
	return nil
}

func (player *OwnerPlayer) sendTile(tile *Tile) {
	if tile.modTime != player.view[tile.Y][tile.X][tile.Z].modTime {
		player.view[tile.Y][tile.X][tile.Z].modTime = tile.modTime
		// NOTE: We could maintain a list of known objects on a tile in the player's tile view and send the difference instead. For large stacks of infrequently changing tiles, this would be more bandwidth efficient, though at the expense of server-side RAM and CPU time.
		// Filter out things we don't want to send to the client.
		filteredMapObjects := make([]ObjectI, 0)
		for _, o := range tile.GetObjects() {
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
							Opaque:      oArch.Matter.Is(cdata.OpaqueMatter),
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

		// Check the given previous knownIDs and see if any were deleted. FIXME: This is kind of inefficient and should probably be handled by the Map.
		for _, oID := range player.view[tile.Y][tile.X][tile.Z].knownIDs {
			if o := tile.gameMap.world.GetObject(oID); o == nil {
				player.ClientConnection.Send(network.CommandObject{
					ObjectID: oID,
					Payload:  network.CommandObjectPayloadDelete{},
				})
			}
		}
		player.view[tile.Y][tile.X][tile.Z].knownIDs = tileObjectIDs
		// Update the client's perception of the given tile.
		player.ClientConnection.Send(network.CommandTile{
			Y:         uint32(tile.Y),
			X:         uint32(tile.X),
			Z:         uint32(tile.Z),
			ObjectIDs: tileObjectIDs,
		})
	}
	if tile.lightModTime != player.view[tile.Y][tile.X][tile.Z].lightModTime {
		player.view[tile.Y][tile.X][tile.Z].lightModTime = tile.lightModTime
		// FIXME: This is a _lot_ of network updates to cause just to update lights... Maybe we should just send the light values of any objects with Light and allow the client to also calculate the brightness and/or r/g/b modulation. This _would_ work fine, however it does mean that clients could just show brightness for a given object if they have seen it once. It won't follow the object, so it might be okay...? It would also give visual bugs if the light object moved out of vision, at least until the source object is found again.
		// Modify by how sky-visibile it is... maybe this should be send to the client directly.
		brightness := tile.sky // * map.brightness..?
		if brightness < 1.0 {
			brightness += tile.brightness
			if brightness > 1 {
				brightness = 1.0
			}
		}
		player.ClientConnection.Send(network.CommandTileLight{
			Y:          uint32(tile.Y),
			X:          uint32(tile.X),
			Z:          uint32(tile.Z),
			Brightness: brightness,
			//Brightness: tile.brightness,
			// TODO: send R, G, B modulation.
		})
	}
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
		if sy = tile.Y - vhh; sy < 0 {
			sy = 0
		}
		if sx = tile.X - vwh; sx < 0 {
			sx = 0
		}
		if sz = tile.Z - vdh; sz < 0 {
			sz = 0
		}
		if ey = tile.Y + vhh; ey > len(player.view) {
			ey = len(player.view)
		}
		if ex = tile.X + vwh; ex > len(player.view[0]) {
			ex = len(player.view[0])
		}
		if ez = tile.Z + vdh; ez > len(player.view[0][0]) {
			ez = len(player.view[0][0])
		}

		for yi := sy; yi < ey; yi++ {
			for xi := sx; xi < ex; xi++ {
				for zi := sz; zi < ez; zi++ {
					player.sendTile(gmap.GetTile(yi, xi, zi))
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
			case OwnerWizardCommand:
				player.wizard = !player.wizard
				player.SendStatus(&StatusWizard{}, player.wizard)
			case OwnerExtCommand:
				if c.Command == "wiz" && player.wizard {
					player.handleWizardCommand(c.Args...)
				}
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

// SendCommand sends the given command to the owner.
func (player *OwnerPlayer) SendCommand(command network.Command) error {
	return player.ClientConnection.Send(command)
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

func (player *OwnerPlayer) handleWizardCommand(args ...string) {
	if len(args) == 0 {
		return
	}
	cmd := args[0]
	args = args[1:]
	switch cmd {
	case "goto":
		mapName := strings.Join(args, " ")
		if gmap, err := player.GetMap().world.LoadMap(mapName); err == nil {
			gmap.AddOwner(player, gmap.y, gmap.x, gmap.z)
		} else {
			log.Printf("Couldn't goto %s: %s\n", mapName, err)
		}
	case "status":
		if len(args) == 0 {
			return
		}
		if status, ok := cdata.StringToStatusMap[args[0]]; ok {
			var s StatusI
			switch status {
			case cdata.FallingStatus:
				s = StatusFallingRef
			case cdata.SqueezingStatus:
				s = StatusSqueezeRef
			case cdata.CrouchingStatus:
				s = StatusCrouchRef
			case cdata.RunningStatus:
				s = StatusRunningRef
			case cdata.SwimmingStatus:
				s = StatusSwimmingRef
			case cdata.FlyingStatus:
				s = StatusFlyingRef
			case cdata.FloatingStatus:
				s = StatusFloatingRef
			}
			if s != nil {
				if player.target.HasStatus(s) {
					player.target.RemoveStatus(s)
				} else {
					player.target.AddStatus(s)
				}
			}
		}
	}
}

// ForgotObject makes the player forget a given object. This will force the object to be resent to the player if it still exists.
func (player *OwnerPlayer) ForgetObject(oID ID) {
	delete(player.knownIDs, oID)
}
