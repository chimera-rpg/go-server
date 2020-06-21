package world

import (
	"github.com/chimera-rpg/go-common/network"

	"fmt"
	"log"
)

type clientConnectionI interface {
	SetOwner(p OwnerI)
	GetOwner() OwnerI
	Send(network.Command) error
	GetID() int
}

// OwnerPlayer represents a player character through a network
// connection and the associated player object.
type OwnerPlayer struct {
	commandChannel   chan OwnerCommand
	ClientConnection clientConnectionI
	target           *ObjectPC
	currentMap       *Map
	mapUpdateTime    uint8
	view             [][][]TileView
	knownIDs         map[ID]struct{}
}

// GetTarget returns the player's target object.
func (player *OwnerPlayer) GetTarget() ObjectI {
	return player.target
}

// SetTarget sets the given object as the target of the player.
func (player *OwnerPlayer) SetTarget(object ObjectI) {
	if objectpc, ok := object.(*ObjectPC); ok {
		player.target = objectpc
	} else {
		log.Printf("Attempted to set OwnerPlayer to non-ObjectPC...\n")
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
	player.CreateView()
}

// NewOwnerPlayer creates a Player from a given client connection.
func NewOwnerPlayer(cc clientConnectionI) *OwnerPlayer {
	return &OwnerPlayer{
		commandChannel:   make(chan OwnerCommand),
		ClientConnection: cc,
		knownIDs:         make(map[ID]struct{}),
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
	if player.mapUpdateTime == player.currentMap.updateTime {
		return
	}
	// Map has changed in some way, so let's check our viewable tiles for updates.
	player.checkVisibleTiles()

	// Make sure we're in sync.
	player.mapUpdateTime = player.currentMap.updateTime
}

// checkVisibleTiles gets the initial view of the player. This sends tile information equal to how far the owner's PC can see.
func (player *OwnerPlayer) checkVisibleTiles() error {
	gmap := player.GetMap()
	// Get owner's viewport.
	vw := 16 // assume 16 for now.
	vh := 16 //
	vd := 16 //
	vwh := vw / 2
	vhh := vh / 2
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
					// TODO: We probably just need to send a TileUpdate that contains an array of object IDs in a given tile. If the player doesn't know what the objectID is, we would send the create payload first.
					player.view[yi][xi][zi].modTime = mapTile.modTime
					for _, o := range mapTile.GetObjects() {
						oID := o.GetID()
						if _, isObjectKnown := player.knownIDs[oID]; !isObjectKnown {
							player.ClientConnection.Send(network.CommandObject{
								ObjectID: o.GetID(),
								Payload: network.CommandObjectPayloadCreate{
									AnimationID: 0,
									FaceID:      0,
									Y:           uint32(yi),
									X:           uint32(xi),
									Z:           uint32(zi),
								},
							})
							player.knownIDs[oID] = struct{}{}
						} else {
							player.ClientConnection.Send(network.CommandObject{
								ObjectID: o.GetID(),
								Payload: network.CommandObjectPayloadMove{
									Y: uint32(yi),
									X: uint32(xi),
									Z: uint32(zi),
								},
							})
						}
					}
				}
			}
		}
	}

	return nil
}

// Update does something.?
func (player *OwnerPlayer) Update(delta int64) error {
	// I guess here is where we'd have some sort of "handleCommandQueue" functionality.
	done := false
	for !done {
		select {
		case pcmd, _ := <-player.commandChannel:
			fmt.Printf("Got owner command: %+v\n", pcmd)
			// Read commands
		default:
			done = true
		}
	}

	return nil
}

// OnMapUpdate is called when the map is updated and the player should update its view and/or react.
func (player *OwnerPlayer) OnMapUpdate(delta int64) error {
	log.Println("checking owner map")
	player.CheckView()

	return nil
}
