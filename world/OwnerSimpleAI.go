package world

import (
	"github.com/chimera-rpg/go-server/data"

	"time"

	log "github.com/sirupsen/logrus"
)

// OwnerSimpleAI represents a non-owner character with a fairly
// simple logic.
type OwnerSimpleAI struct {
	target        *ObjectCharacter
	currentMap    *Map
	mapUpdateTime uint8
	knownIDs      map[ID]struct{}
	attitudes     map[ID]data.Attitude
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
		knownIDs: make(map[ID]struct{}),
	}
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
