package world

import (
	"math/rand"
	"time"

	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
)

// ObjectFlora represents most non-NPC, non-PC, and non-animal living things within the game.
type ObjectFlora struct {
	Object
	//
	elapsedGrowth time.Duration
	nextGrowth    time.Duration
	nextArchetype *data.Archetype
}

// NewObjectFlora returns an ObjectFlora from the given Archetype.
func NewObjectFlora(a *data.Archetype) (o *ObjectFlora) {
	o = &ObjectFlora{
		Object: NewObject(a),
	}
	o.getGrowth(a)

	return
}

func (o *ObjectFlora) getGrowth(a *data.Archetype) {
	o.updates = false
	o.nextArchetype = nil
	if a.Next.MaxTime != 0 && a.Next.MinTime != 0 && len(a.Next.Archetypes) > 0 {
		if a.Next.MinTime == a.Next.MaxTime {
			o.nextGrowth = time.Duration(a.Next.MaxTime) * time.Second
		} else {
			o.nextGrowth = time.Duration(rand.Intn(int(a.Next.MaxTime)-int(a.Next.MinTime))+int(a.Next.MinTime)) * time.Second
		}
		// Let's just roll the next archetype now.
		sum := 0.0
		for _, next := range a.Next.Archetypes {
			sum += float64(next.Weight)
		}
		nextRand := rand.Float64() * sum
		for _, archetype := range a.Next.Archetypes {
			if nextRand < float64(archetype.Weight) {
				o.nextArchetype = &archetype.Archetype
				break
			}
			nextRand -= float64(archetype.Weight)
		}
		if o.nextArchetype != nil {
			o.updates = true
		}
	}
}

func (o *ObjectFlora) update(delta time.Duration) {
	o.elapsedGrowth += delta
	if o.elapsedGrowth >= o.nextGrowth {
		o.Archetype = o.nextArchetype
		o.blocking = o.nextArchetype.Blocking

		// Refresh it.
		o.tile.gameMap.RefreshObject(o.id)

		// Get next stage, if applicable.
		o.getGrowth(o.nextArchetype)

		// Remove from the thinkin' map if we have nothing to grow into.
		if o.nextArchetype == nil {
			o.tile.gameMap.InactiveObject(o.id)
		}
	}
}

func (o *ObjectFlora) getType() cdata.ArchetypeType {
	return cdata.ArchetypeFlora
}
