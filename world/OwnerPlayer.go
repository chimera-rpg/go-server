package world

import (
	"strings"

	"github.com/chimera-rpg/go-server/data"
	"github.com/chimera-rpg/go-server/network"

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

type dummyConnection struct {
	owner *OwnerPlayer
	user  *data.User
	id    int
}

func (c *dummyConnection) GetUser() *data.User {
	return c.user
}
func (c *dummyConnection) SetOwner(p *OwnerPlayer) {
	c.owner = p
}
func (c *dummyConnection) GetOwner() *OwnerPlayer {
	return c.owner
}
func (c *dummyConnection) Send(network.Command) error {
	return nil
}
func (c *dummyConnection) GetID() int {
	return c.id
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
	disconnected                     bool
	disconnectedElapsed              time.Duration
}

// GetTarget returns the player's target object.
func (player *OwnerPlayer) GetTarget() ObjectI {
	return player.target
}

func (player *OwnerPlayer) HasDummyConnection() bool {
	_, ok := player.ClientConnection.(*dummyConnection)
	return ok
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
			Name:         m.name,
			MapID:        m.mapID,
			Height:       m.height,
			Width:        m.width,
			Depth:        m.depth,
			AmbientRed:   m.ambientRed,
			AmbientGreen: m.ambientGreen,
			AmbientBlue:  m.ambientBlue,
			Outdoor:      m.outdoor,
			OutdoorRed:   m.outdoorRed,
			OutdoorGreen: m.outdoorGreen,
			OutdoorBlue:  m.outdoorBlue,
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
	player.checkVisionRing()
}

// SetViewSize sets the viewport limits of the player.
func (player *OwnerPlayer) SetViewSize(h, w, d int) {
	player.viewHeight = h
	player.viewWidth = w
	player.viewDepth = d
	// When our view size changes, send it to the client.
	player.ClientConnection.Send(network.CommandObject{
		ObjectID: player.GetTarget().GetID(),
		Payload: network.CommandObjectPayloadViewTarget{
			Height: uint8(player.viewHeight),
			Width:  uint8(player.viewWidth),
			Depth:  uint8(player.viewDepth),
		},
	})
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

// getVisionCube2 acquires a hollow cube.
func (player *OwnerPlayer) getVisionCube2() (c [][3]int) {
	tile := player.GetTarget().GetTile()
	m := tile.GetMap()
	vh, vw, vd := player.GetViewSize()
	vhh := vh / 2
	vwh := vw / 2
	vdh := vd / 2
	y1 := tile.Y + int(player.GetTarget().GetArchetype().Height) - 1

	ymin := y1 - vhh
	if ymin < 0 {
		ymin = 0
	}
	ymax := y1 + vhh
	if ymax > m.height {
		ymax = m.height
	}

	xmin := tile.X - vwh
	if xmin < 0 {
		xmin = 0
	}
	xmax := tile.X + vwh
	if xmax > m.width {
		xmax = m.width
	}

	zmin := tile.Z - vdh
	if zmin < 0 {
		zmin = 0
	}
	zmax := tile.Z + vdh
	if zmax > m.depth {
		zmax = m.depth
	}

	for y := ymin; y < ymax; y++ {
		for x := xmin; x < xmax; x++ {
			for z := zmin; z < zmax; z++ {
				if (y == ymin || y == ymax-1) || (x == xmin || x == xmax-1) || (z == zmin || z == zmax-1) {
					c = append(c, [3]int{y, x, z})
				}
			}
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
		ymax = m.height
	}

	xmin := tile.X - vwh
	if xmin < 0 {
		xmin = 0
	}
	xmax := tile.X + vwh
	if xmax > m.width {
		xmax = m.width
	}

	zmin := tile.Z - vdh
	if zmin < 0 {
		zmin = 0
	}
	zmax := tile.Z + vdh
	if zmax > m.depth {
		zmax = m.depth
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
	var tiles []*Tile
	var tileUpdates []network.CommandTile
	var skyUpdates []network.CommandTileSky
	var lightUpdates []network.CommandTileLight

	gmap := player.GetMap()
	tile := player.GetTarget().GetTile()
	//coords := player.getVisionRing()
	coords := player.getVisionCube2()

	// Ensure our own tile is updated.
	if player.checkTile(tile) {
		tileUpdates = append(tileUpdates, network.CommandTile{
			Y:         uint32(tile.Y),
			X:         uint32(tile.X),
			Z:         uint32(tile.Z),
			ObjectIDs: player.view[tile.Y][tile.X][tile.Z].knownIDs,
		})
	}

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
		tiles = append(tiles, gmap.ShootRay(y1, x1, z1, float64(c[0]), float64(c[1]), float64(c[2]), func(t *Tile) bool {
			return !t.opaque
		})...)
	}

	var hasUpdates bool
	for _, tile := range tiles {
		if player.checkTile(tile) {
			tileUpdates = append(tileUpdates, network.CommandTile{
				Y:         uint32(tile.Y),
				X:         uint32(tile.X),
				Z:         uint32(tile.Z),
				ObjectIDs: player.view[tile.Y][tile.X][tile.Z].knownIDs,
			})
			hasUpdates = true
		}
		if tile.skyModTime != player.view[tile.Y][tile.X][tile.Z].skyModTime {
			player.view[tile.Y][tile.X][tile.Z].skyModTime = tile.skyModTime
			skyUpdates = append(skyUpdates, network.CommandTileSky{
				Y:   uint32(tile.Y),
				X:   uint32(tile.X),
				Z:   uint32(tile.Z),
				Sky: float64(tile.sky),
			})
			hasUpdates = true
		}
		if tile.lightModTime != player.view[tile.Y][tile.X][tile.Z].lightModTime {
			player.view[tile.Y][tile.X][tile.Z].lightModTime = tile.lightModTime
			lightUpdates = append(lightUpdates, network.CommandTileLight{
				Y: uint32(tile.Y),
				X: uint32(tile.X),
				Z: uint32(tile.Z),
				R: tile.r,
				G: tile.g,
				B: tile.b,
			})
			hasUpdates = true
		}
	}
	batchUpdates := false
	if hasUpdates {
		if batchUpdates {
			player.ClientConnection.Send(network.CommandTiles{
				TileUpdates:  tileUpdates,
				LightUpdates: lightUpdates,
				SkyUpdates:   skyUpdates,
			})
		} else {
			for _, t := range tileUpdates {
				player.ClientConnection.Send(t)
			}
			for _, t := range lightUpdates {
				player.ClientConnection.Send(t)
			}
			for _, t := range skyUpdates {
				player.ClientConnection.Send(t)
			}
		}
	}

	return nil
}

func (player *OwnerPlayer) sendObject(o ObjectI) {
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
					Reach:       oArch.Reach,
					Opaque:      oArch.Matter.Is(data.OpaqueMatter),
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
}

func (player *OwnerPlayer) checkTile(tile *Tile) bool {
	if tile.modTime != player.view[tile.Y][tile.X][tile.Z].modTime {
		player.view[tile.Y][tile.X][tile.Z].modTime = tile.modTime
		// NOTE: We could maintain a list of known objects on a tile in the player's tile view and send the difference instead. For large stacks of infrequently changing tiles, this would be more bandwidth efficient, though at the expense of server-side RAM and CPU time.
		// Filter out things we don't want to send to the client.
		filteredMapObjects := make([]ObjectI, 0)
		for _, o := range tile.GetObjects() {
			if o.getType() != data.ArchetypeAudio && o.getType() != data.ArchetypeSpecial {
				filteredMapObjects = append(filteredMapObjects, o)
			}
		}
		tileObjectIDs := make([]ID, len(filteredMapObjects))
		// Send any objects unknown to the client (and collect their IDs).
		for i, o := range filteredMapObjects {
			player.sendObject(o)
			tileObjectIDs[i] = o.GetID()
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
		return true
	}
	return false
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
		if status, ok := data.StringToStatusMap[args[0]]; ok {
			var s StatusI
			switch status {
			case data.FallingStatus:
				s = StatusFallingRef
			case data.SqueezingStatus:
				s = StatusSqueezeRef
			case data.CrouchingStatus:
				s = StatusCrouchRef
			case data.RunningStatus:
				s = StatusRunningRef
			case data.SwimmingStatus:
				s = StatusSwimmingRef
			case data.FlyingStatus:
				s = StatusFlyingRef
			case data.FloatingStatus:
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
