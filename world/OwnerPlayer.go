package world

import (
	"math"

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
	y1 := tile.y + int(player.GetTarget().GetArchetype().Height)
	// TODO: Use target object's statistics for vision range.
	add := func(y, x, z int) {
		if y < 0 || x < 0 || z < 0 || y >= m.height || x >= m.width || z >= m.depth {
			return
		}
		c = append(c, [3]int{y, x, z})
	}
	// bottom & top
	for x := tile.x - vwh; x < tile.x+vwh; x++ {
		for z := tile.z - vdh; z < tile.z+vdh; z++ {
			add(y1-vhh, x, z)
			add(y1+vhh, x, z)
		}
	}
	// left & right
	for y := tile.y - vhh; y < tile.y+vhh; y++ {
		for z := tile.z - vdh; z < tile.z+vdh; z++ {
			add(y, tile.x-vwh, z)
			add(y, tile.x+vwh, z)
		}
	}
	// back & front
	for y := tile.y - vhh; y < tile.y+vhh; y++ {
		for x := tile.x - vwh; x < tile.x+vwh; x++ {
			add(y, x, tile.z-vdh)
			add(y, x, tile.z+vdh)
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
	y1 := tile.y + int(player.GetTarget().GetArchetype().Height)

	ymin := y1 - vhh
	if ymin < 0 {
		ymin = 0
	}
	ymax := y1 + vhh
	if ymax > m.height {
		ymax = m.height - 1
	}

	xmin := tile.x - vwh
	if xmin < 0 {
		xmin = 0
	}
	xmax := tile.x + vwh
	if xmax > m.width {
		xmax = m.width - 1
	}

	zmin := tile.z - vdh
	if zmin < 0 {
		zmin = 0
	}
	zmax := tile.z + vdh
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
	y1 := float64(tile.y + int(a.Height))
	if y1 >= float64(gmap.height) {
		y1 = float64(gmap.height - 1)
	}
	x1 := float64(tile.x) + float64(a.Width)/2
	z1 := float64(tile.z) + float64(a.Depth)/2
	for _, c := range coords {
		var tMaxX, tMaxY, tMaxZ, tDeltaX, tDeltaY, tDeltaZ float64
		y2 := float64(c[0])
		x2 := float64(c[1])
		z2 := float64(c[2])
		var dy, dx, dz int
		var y, x, z int

		sign := func(x float64) int {
			if x > 0 {
				return 1
			} else if x < 0 {
				return -1
			}
			return 0
		}
		frac0 := func(x float64) float64 {
			return x - math.Floor(x)
		}
		frac1 := func(x float64) float64 {
			return 1 - x + math.Floor(x)
		}

		dy = sign(y2 - y1)
		if dy != 0 {
			tDeltaY = math.Min(float64(dy)/(y2-y1), 10000)
		} else {
			tDeltaY = 10000
		}
		if dy > 0 {
			tMaxY = tDeltaY * frac1(y1)
		} else {
			tMaxY = tDeltaY * frac0(y1)
		}
		y = int(y1)

		dx = sign(x2 - x1)
		if dx != 0 {
			tDeltaX = math.Min(float64(dx)/(x2-x1), 10000)
		} else {
			tDeltaX = 10000
		}
		if dx > 0 {
			tMaxX = tDeltaX * frac1(x1)
		} else {
			tMaxX = tDeltaX * frac0(x1)
		}
		x = int(x1)

		dz = sign(z2 - z1)
		if dz != 0 {
			tDeltaZ = math.Min(float64(dz)/(z2-z1), 10000)
		} else {
			tDeltaZ = 10000
		}
		if dz > 0 {
			tMaxZ = tDeltaZ * frac1(z1)
		} else {
			tMaxZ = tDeltaZ * frac0(z1)
		}
		z = int(z1)

		for {
			if tMaxX < tMaxY {
				if tMaxX < tMaxZ {
					x += dx
					tMaxX += tDeltaX
				} else {
					z += dz
					tMaxZ += tDeltaZ
				}
			} else {
				if tMaxY < tMaxZ {
					y += dy
					tMaxY += tDeltaY
				} else {
					z += dz
					tMaxZ += tDeltaZ
				}
			}
			if tMaxY > 1 && tMaxX > 1 && tMaxZ > 1 {
				break
			}
			if y < 0 || x < 0 || z < 0 || y >= gmap.height || x >= gmap.width || z >= gmap.depth {
				continue
				//break
			}
			tile := gmap.GetTile(y, x, z)
			player.sendTile(tile)
			if tile.opaque {
				break
			}
		}
	}
	return nil
}

func (player *OwnerPlayer) sendTile(tile *Tile) {
	// TODO: Actually calculate which tiles are visible for the owner.
	if tile.modTime == player.view[tile.y][tile.x][tile.z].modTime {
		return
	}
	player.view[tile.y][tile.x][tile.z].modTime = tile.modTime
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
	// Update the client's perception of the given tile.
	player.ClientConnection.Send(network.CommandTile{
		Y:         uint32(tile.y),
		X:         uint32(tile.x),
		Z:         uint32(tile.z),
		ObjectIDs: tileObjectIDs,
	})
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

func (player *OwnerPlayer) handleWizardCommand(args ...string) {
	if len(args) == 0 {
		return
	}
	cmd := args[0]
	args = args[1:]
	switch cmd {
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
