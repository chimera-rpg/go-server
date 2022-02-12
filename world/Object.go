package world

import (
	"math"
	"math/rand"
	"reflect"
	"time"

	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
)

// Object is the base type that should be used as an embeded struct in
// all game Objects.
type Object struct {
	Archetype    *data.Archetype
	AltArchetype *data.Archetype
	id           ID
	// Relationships
	tile   *Tile
	parent ObjectI
	owner  OwnerI
	//
	statuses   []StatusI
	inventory  []ObjectI
	hasMoved   bool
	updates    bool
	blocking   cdata.MatterType
	actions    int // Actions are the amount of actions that a player can take within 1 second
	maxActions int // Max actions are the amount of actions that a player can take within 1 second.
	//
	timers []Timer
}

// NewObject returns a new Object that references the given Archetype.
func NewObject(a *data.Archetype) Object {
	o := Object{
		blocking:  a.Blocking,
		Archetype: a,
	}

	o.addTimers(a.Timers)

	return o
}

// update updates the given object.
func (o *Object) update(delta time.Duration) {
	// Handle timers.
	if len(o.timers) > 0 {
		temp := o.timers[:0]
		for _, t := range o.timers {
			t.elapsed += delta
			if t.elapsed >= t.target {
				// Process
				t.elapsed -= t.target

				if t.repeat <= t.repeatCount {
					temp = append(temp, t)
				}
				t.repeatCount++
				// Trigger associated event.
				switch t.event {
				case "Birth":
					o.ResolveEvent(EventBirth{})
				case "Advance":
					o.ResolveEvent(EventAdvance{})
				}
			} else {
				temp = append(temp, t)
			}
		}
		o.timers = temp
	}

	for i := 0; i < len(o.statuses); i++ {
		o.statuses[i].update(delta)
		if o.statuses[i].ShouldRemove() {
			if o.RemoveStatus(o.statuses[i]) != nil {
				i--
			}
		}
	}
}

// GetOwner returns the owning object.
func (o *Object) GetOwner() OwnerI {
	return o.owner
}

// SetOwner sets the owner to the given object.
func (o *Object) SetOwner(owner OwnerI) {
	// TODO: check if owner is set
	o.owner = owner
}

// SetTile sets the tile to the given tile.
func (o *Object) SetTile(tile *Tile) {
	o.tile = tile
}

// GetTile gets the tile where the object is contained.
func (o *Object) GetTile() *Tile {
	return o.tile
}

// SetMoved sets the object's hasMoved to the given value.
func (o *Object) SetMoved(b bool) {
	o.hasMoved = b
}

// SetID sets the objects' id. This should _only_ be called by World during object creation.
func (o *Object) SetID(id ID) {
	o.id = id
}

// GetID gets the object's id.
func (o *Object) GetID() ID {
	return o.id
}

func (o *Object) setArchetype(targetArch *data.Archetype) {
}

// GetArchetype gets the object's underlying archetype.
func (o *Object) GetArchetype() *data.Archetype {
	return o.Archetype
}

// GetAltArchetype gets the object's underlying alt archetype. This is used for ObjectCharacters to store an uncompiled version of their archetype that can be easily saved to disk.
func (o *Object) GetAltArchetype() *data.Archetype {
	return o.AltArchetype
}

func (o *Object) getType() cdata.ArchetypeType {
	return cdata.ArchetypeUnknown
}

// AddStatus adds the given status to the object.
func (o *Object) AddStatus(s StatusI) {
	s.SetTarget(o)
	o.statuses = append(o.statuses, s)
	s.OnAdd()
}

// RemoveStatus removes the given status from the object, returning the status that was stored.
func (o *Object) RemoveStatus(s StatusI) StatusI {
	for i, s2 := range o.statuses {
		if reflect.TypeOf(s) == reflect.TypeOf(s2) {
			o.statuses = append(o.statuses[:i], o.statuses[i+1:]...)
			s2.OnRemove()
			if o.GetOwner() != nil {
				o.GetOwner().SendStatus(s2, false)
			}
			return s2
		}
	}
	return nil
}

// HasStatus checks if the object has the given status.
func (o *Object) HasStatus(t StatusI) bool {
	for _, s := range o.statuses {
		if reflect.TypeOf(s) == reflect.TypeOf(t) {
			return true
		}
	}
	return false
}

// SetStatus sets the status.
func (o *Object) SetStatus(t StatusI) bool {
	return false
}

// ResolveEvent is the default handler for object events.
func (o *Object) ResolveEvent(e EventI) bool {
	if o.Archetype != nil && o.Archetype.Events != nil {
		events := o.Archetype.Events
		switch e.(type) {
		case EventBirth:
			if events.Birth != nil {
				o.processEventResponses(events.Birth)
			}
		case EventAdvance:
			if events.Advance != nil {
				o.processEventResponses(events.Advance)
			}
		}
	}
	// Do nothing per default.
	return true
}

func (o *Object) processEventResponses(r *data.EventResponses) {
	// Replace replaces the given object's underline archetype with a randomly weighted one. Note that replace removes other running timers!
	if r.Replace != nil {
		// Do nothing if we have no actual archetype list to use.
		if len(*r.Replace) == 0 {
			return
		}
		var archetype *data.Archetype
		sum := 0.0
		for _, a := range *r.Replace {
			sum += float64(a.Chance)
		}
		// If sum is zero, assign archetype to the first index entry.
		if sum == 0 {
			archetype = (*r.Replace)[0].Archetype
		} else {
			nextRand := rand.Float64() * sum
			for _, a := range *r.Replace {
				if nextRand < float64(a.Chance) {
					archetype = a.Archetype
					break
				}
				nextRand -= float64(a.Chance)
			}
		}

		// We got an archetype, let's replace.
		if archetype != nil {
			o.replaceArchetype(archetype)
		}
	}
	/*if r.Spawn != nil {
	}
	if r.Trigger != nil {
	}*/
}

func (o *Object) replaceArchetype(a *data.Archetype) {
	o.Archetype = a
	o.blocking = o.Archetype.Blocking

	// Force a refresh.
	o.tile.gameMap.RefreshObject(o.id)

	// Clear old timers.
	o.timers = make([]Timer, 0)
	// Inactive object if we have no timers.
	if len(o.Archetype.Timers) == 0 {
		o.tile.gameMap.InactiveObject(o.id)
	} else {
		// Otherwise add timers!
		o.addTimers(o.Archetype.Timers)
	}
}

func (o *Object) addTimers(timers []data.ArchetypeTimer) {
	for _, t := range timers {
		o.timers = append(o.timers,
			Timer{
				elapsed:     0,
				target:      t.Wait.Random(),
				event:       t.Event,
				repeat:      t.Repeat,
				repeatCount: 0,
			},
		)
	}
}

// Timers returns the object's timers.
func (o *Object) Timers() *[]Timer {
	return &o.timers
}

// GetStatus returns the associated status.
func (o *Object) GetStatus(t StatusI) StatusI {
	for _, s := range o.statuses {
		if reflect.TypeOf(s) == reflect.TypeOf(t) {
			return s
		}
	}
	return nil
}

// Blocks returns if the object blocks the given MatterType.
func (o *Object) Blocks(matter cdata.MatterType) bool {
	return o.blocking.Is(matter)
}

// Matter returns the object's matter, acquired from its archetype.
func (o *Object) Matter() cdata.MatterType {
	a := o.GetArchetype()
	return a.Matter
}

// Name returns the name of the object, if available.
func (o *Object) Name() string {
	return ""
}

func (o *Object) GetDimensions() (h, w, d int) {
	a := o.GetArchetype()
	if a != nil {
		h = int(a.Height)
		w = int(a.Width)
		d = int(a.Depth)
	}
	if s := o.GetStatus(&StatusSqueeze{}); s != nil {
		t := s.(*StatusSqueeze)
		w -= t.X
		d -= t.Z
	}
	if s := o.GetStatus(&StatusCrouch{}); s != nil {
		t := s.(*StatusCrouch)
		h -= t.Y
	}
	return
}

// Stamina returns the object's stamina.
func (o *Object) Stamina() int {
	return 0
}

// MaxStamina returns the object's maximum stamina.
func (o *Object) MaxStamina() int {
	return 0
}

// RestoreStamina DOES NOTHING
func (o *Object) RestoreStamina() {}

// GetDistance gets the distance from the object to the target coordinates.
func (o *Object) GetDistance(y, x, z int) float64 {
	t := o.GetTile()
	return math.Sqrt(math.Pow(float64(y-t.y), 2) + math.Pow(float64(x-t.x), 2) + math.Pow(float64(z-t.z), 2))
}

func (o *Object) Updates() bool {
	return o.updates
}
