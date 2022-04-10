package world

import (
	"github.com/chimera-rpg/go-common/network"
	"github.com/chimera-rpg/go-server/data"

	"time"

	log "github.com/sirupsen/logrus"
)

// OwnerSimpleAI represents a non-owner character with a fairly
// simple logic.
type OwnerSimpleAI struct {
	Owner
	target                           *ObjectCharacter
	currentMap                       *Map
	mapUpdateTime                    uint8
	knownIDs                         map[ID]struct{}
	attitudes                        map[ID]data.Attitude
	viewHeight, viewWidth, viewDepth int
}

// GetTarget returns the owners's target object.
func (owner *OwnerSimpleAI) GetTarget() ObjectI {
	return owner.target
}

// SetTarget sets the given object as the target of the owner.
func (owner *OwnerSimpleAI) SetTarget(object ObjectI) {
	if objectnpc, ok := object.(*ObjectCharacter); ok {
		owner.target = objectnpc
	} else {
		log.Printf("Attempted to set OwnerSimpleAI to non-ObjectCharacter...\n")
	}
	object.SetOwner(owner)
}

// GetMap gets the currentMap of the owner.
func (owner *OwnerSimpleAI) GetMap() *Map {
	return owner.currentMap
}

// SetMap sets the currentMap of the owner.
func (owner *OwnerSimpleAI) SetMap(m *Map) {
	owner.currentMap = m
}

// NewOwnerSimpleAI creates a new OwnerSimpleAI.
func NewOwnerSimpleAI() *OwnerSimpleAI {
	return &OwnerSimpleAI{
		knownIDs:   make(map[ID]struct{}),
		viewHeight: 8,
		viewWidth:  16,
		viewDepth:  16,
	}
}

// SetViewSize sets the viewport limits of the player.
func (owner *OwnerSimpleAI) SetViewSize(h, w, d int) {
	owner.viewHeight = h
	owner.viewWidth = w
	owner.viewDepth = d
}

// GetViewSize returns the view port size that is used to send map updates to the player.
func (owner *OwnerSimpleAI) GetViewSize() (h, w, d int) {
	// TODO: Possibly replace with target object's vision.
	return owner.viewHeight, owner.viewWidth, owner.viewDepth
}

// Update does something.?
func (owner *OwnerSimpleAI) Update(delta time.Duration) error {
	// TODO: Handle a state machine or similar here.
	return nil
}

// OnMapUpdate is called when the map is updated.
func (owner *OwnerSimpleAI) OnMapUpdate(delta time.Duration) error {
	if owner.mapUpdateTime == owner.currentMap.updateTime {
		return nil
	}

	// TODO: Set some sort of flag for the AI to check its view on next Update.

	// Make sure we're in sync.
	owner.mapUpdateTime = owner.currentMap.updateTime

	return nil
}

// OnObjectDelete is called when an object on the map is deleted. If the owner knows about it, then an object delete command is sent to the client.
func (owner *OwnerSimpleAI) OnObjectDelete(oID ID) error {
	if _, isObjectKnown := owner.knownIDs[oID]; isObjectKnown {
		delete(owner.knownIDs, oID)
	}
	return nil
}

// GetAttitude returns the attitude the owner has the a given object. If no attitude exists, one is calculated based upon the target's attitude (if it has one).
func (owner *OwnerSimpleAI) GetAttitude(oID ID) data.Attitude {
	if attitude, ok := owner.attitudes[oID]; ok {
		return attitude
	}
	target := owner.GetMap().world.GetObject(oID)
	if target == nil {
		delete(owner.attitudes, oID)
	} else {
		// TODO: We should probably check if the target knows us and use their attitude. If not, we should calculate from our target object archetype's default attitude towards: Genera, Species, Legacy, and Faction.
		if otherOwner := target.GetOwner(); otherOwner != nil {
			return otherOwner.GetAttitude(owner.target.id)
		}
	}
	return data.NoAttitude
}

// SendCommand sends the given command to the owner.
func (owner *OwnerSimpleAI) SendCommand(command network.Command) error {
	return nil
}

// SendMessage sends a message to the character.
func (owner *OwnerSimpleAI) SendMessage(s string) {
}

// SendStatus sends the status to the owner, providing it has a StatusType.
func (owner *OwnerSimpleAI) SendStatus(s StatusI, active bool) {
}

// SendSound does nothing.
func (owner *OwnerSimpleAI) SendSound(audioID, soundID ID, objectID ID, y, x, z int, volume float32) {
}

// SendMusic does nothing.
func (owner *OwnerSimpleAI) SendMusic(audioID, soundID ID, soundIndex int8, objectID ID, y, x, z int, volume float32, loopCount int8) {
}

// StopMusic does nothing.
func (owner *OwnerSimpleAI) StopMusic(objectID ID) {
}

// ShootRay shoots out a ray and returns all tiles between the starting position and the target coordinate.
func (owner *OwnerSimpleAI) ShootRay(y, x, z float64, f func(tile *Tile) bool) (tiles []*Tile) {
	gmap := owner.GetMap()
	t := owner.target.tile
	h, w, d := owner.target.GetDimensions()
	return gmap.ShootRay(float64(t.Y)+float64(h)/2, float64(t.X)+float64(w)/2, float64(t.Z)+float64(d/2), y, x, z, f)
}
