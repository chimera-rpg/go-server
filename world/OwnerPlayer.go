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
	view             [][][]TileView
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

// SendCommand
func (player *OwnerPlayer) SendCommand(cmd network.Command) error {
	return player.ClientConnection.Send(cmd)
}

// GetMap gets the currentMap of the owner.
func (player *OwnerPlayer) GetMap() *Map {
	return player.currentMap
}

// SetMap sets the currentMap of the owner.
func (player *OwnerPlayer) SetMap(m *Map) {
	player.currentMap = m
}

// NewOwnerPlayer creates a Player from a given client connection.
func NewOwnerPlayer(cc clientConnectionI) *OwnerPlayer {
	return &OwnerPlayer{
		commandChannel:   make(chan OwnerCommand),
		ClientConnection: cc,
	}
}

// CreateView creates the initial view of the player.
func (player *OwnerPlayer) CreateView() {
	vw := 16 // assume 16 for now.
	vh := 16 //
	vd := 16 //
	player.view = make([][][]TileView, vh)
	for y := 0; y < len(player.view); y++ {
		player.view[y] = make([][]TileView, vw)
		for x := 0; x < len(player.view[y]); x++ {
			player.view[y][x] = make([]TileView, vd)
		}
	}
}

// AcquireView gets the initial view of the player. This sends tile information equal to how far the owner's PC can see.
func (player *OwnerPlayer) AcquireView() error {
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
		if ey = tile.y + vhh; ey > gmap.height {
			ey = gmap.height
		}
		if ex = tile.x + vwh; ex > gmap.width {
			ex = gmap.width
		}
		if ez = tile.z + vdh; ez > gmap.depth {
			ez = gmap.depth
		}

		for yi := sy; yi < ey; yi++ {
			for xi := sx; xi < ex; xi++ {
				for zi := sz; zi < ez; zi++ {
					// TODO: Actually calculate which tiles are visible for the owner.
					for _, o := range gmap.GetTile(yi, xi, zi).GetObjects() {
						player.SendCommand(network.CommandObject{
							ObjectID: o.GetID(),
							Payload: network.CommandObjectPayloadCreate{
								AnimationID: 0,
								FaceID:      0,
								Y:           uint32(yi),
								X:           uint32(xi),
								Z:           uint32(zi),
							},
						})
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
