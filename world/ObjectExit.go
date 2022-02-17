package world

import (
	"errors"

	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
)

// ObjectExit represents entrance/exit/teleporter objects.
type ObjectExit struct {
	Object
	cooldown int
}

// NewObjectExit returns an ObjectExit from the given Archetype.
func NewObjectExit(a *data.Archetype) (o *ObjectExit) {
	o = &ObjectExit{
		Object: NewObject(a),
	}
	return
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
	return nil
}

// IsReady returns if the exit is ready for use (its cooldown is greater/equal to its arch Cooldown value).
func (o *ObjectExit) IsReady() bool {
	return o.cooldown >= int(o.Archetype.Exit.Cooldown)
}

func (o *ObjectExit) getType() cdata.ArchetypeType {
	return cdata.ArchetypeExit
}
