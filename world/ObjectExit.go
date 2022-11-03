package world

import (
	"errors"
	"fmt"
	"time"

	"github.com/chimera-rpg/go-server/data"
)

// ObjectExit represents entrance/exit/teleporter objects.
type ObjectExit struct {
	Object
	cooldown   data.Duration
	uses       int
	uniqueUses map[uint32]int
}

// NewObjectExit returns an ObjectExit from the given Archetype.
func NewObjectExit(a *data.Archetype) (o *ObjectExit) {
	o = &ObjectExit{
		Object: NewObject(a),
	}
	if a.Exit != nil {
		o.cooldown = a.Exit.Cooldown
		if a.Exit.UniqueUses > 0 {
			o.uniqueUses = make(map[uint32]int)
		}
	}
	return
}

func (o *ObjectExit) Updates() bool {
	return o.cooldown.Duration > 0
}

func (o *ObjectExit) update(delta time.Duration) {
	// Inactivate the exit object if its cooldown has reduced.
	o.cooldown.Duration += delta
	fmt.Println(o.cooldown, o.Archetype.Exit.Cooldown.Duration)
	if o.cooldown.Duration >= o.Archetype.Exit.Cooldown.Duration {
		o.tile.gameMap.InactiveObject(o.id)
	}
}

// Teleport moves the target object based upon the rules of the exit archetype. Returns nil if the teleport was successful or an error on failure.
func (o *ObjectExit) Teleport(target ObjectI) error {
	if !o.IsReady() {
		// TODO: Probably a cooldown message?
		return errors.New("not ready")
	}
	if o.Archetype.Exit == nil {
		return errors.New("nil exit")
	}
	if o.Archetype.Exit.Uses > 0 && o.uses >= o.Archetype.Exit.Uses {
		return errors.New("no more uses")
	}
	if o.Archetype.Exit.UniqueUses > 0 {
		if uses, ok := o.uniqueUses[target.GetID()]; ok {
			if uses >= o.Archetype.Exit.UniqueUses {
				return errors.New("no more unique uses")
			}
		}
	}
	// Check if the target object is large enough to trigger/use the exit.
	if o.Archetype.Exit.SizeRatio > 0 && o.Archetype.Exit.SizeRatio < 1 {
		h, w, d := o.GetDimensions()
		t := float64(h + w + d)
		th, tw, td := target.GetDimensions()
		t2 := float64(th + tw + td)
		r := t / t2
		if r < o.Archetype.Exit.SizeRatio {
			return errors.New("too large")
		}
	}

	// Scripting check
	if o.Archetype.Events != nil && o.Archetype.Events.Exit != nil && o.Archetype.Events.Exit.Script != nil {
		e := EventExit{
			Target:  target,
			Prevent: false,
			Message: "prevented by script",
		}
		o.processEventResponses(o.Archetype.Events.Exit, &e)
		if e.Prevent {
			return errors.New(e.Message)
		}
	}

	if o.Archetype.Exit.Name == "" { // Same map teleport.
		y := o.tile.GetMap().y
		x := o.tile.GetMap().x
		z := o.tile.GetMap().z
		if o.Archetype.Exit.Y != nil {
			y = *o.Archetype.Exit.Y
		}
		if o.Archetype.Exit.X != nil {
			x = *o.Archetype.Exit.X
		}
		if o.Archetype.Exit.Z != nil {
			z = *o.Archetype.Exit.Z
		}
		return o.GetTile().GetMap().TeleportObject(target, y, x, z, true)
	} else { // Other map.
		// Only move character objects between maps. NOTE: We could allow teleporting objects between maps here!
		if target, isCharacter := target.(*ObjectCharacter); isCharacter {
			if gmap, err := o.GetTile().GetMap().world.LoadMap(o.Archetype.Exit.Name); err == nil { // Travel archetype to map.
				y := gmap.y
				x := gmap.x
				z := gmap.z
				if o.Archetype.Exit.Y != nil {
					y = *o.Archetype.Exit.Y
				}
				if o.Archetype.Exit.X != nil {
					x = *o.Archetype.Exit.X
				}
				if o.Archetype.Exit.Z != nil {
					z = *o.Archetype.Exit.Z
				}
				gmap.AddOwner(target.GetOwner(), y, x, z)
			} else { // Map failed to load!
				return err
			}
		}
	}
	// If the exit has a cooldown, make the object active.
	if o.Archetype.Exit.Cooldown.Duration > 0 {
		o.cooldown.Duration = 0
		o.tile.gameMap.ActivateObject(o.id)
	}
	o.uses++
	if o.Archetype.Exit.UniqueUses > 0 {
		if uses, ok := o.uniqueUses[target.GetID()]; ok {
			o.uniqueUses[target.GetID()] = uses + 1
		} else {
			o.uniqueUses[target.GetID()] = 1
		}
	}
	return nil
}

// IsReady returns if the exit is ready for use (its cooldown is greater/equal to its arch Cooldown value).
func (o *ObjectExit) IsReady() bool {
	return o.cooldown.Duration >= o.Archetype.Exit.Cooldown.Duration
}

func (o *ObjectExit) getType() data.ArchetypeType {
	return data.ArchetypeExit
}
