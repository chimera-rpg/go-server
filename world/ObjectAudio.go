package world

import (
	"strconv"
	"time"

	"github.com/chimera-rpg/go-server/data"
)

// ObjectAudio represents audio for in-map sound and music playback.
type ObjectAudio struct {
	Object
	loopCount int8
	volume    float32
	elapsed   time.Duration
	listeners map[ID]struct{}
}

// NewObjectAudio creates a sound and music playback object from the given archetype.
func NewObjectAudio(a *data.Archetype) (o *ObjectAudio) {
	o = &ObjectAudio{
		Object:    NewObject(a),
		listeners: make(map[ID]struct{}),
	}
	// Use count for loops and value for volume.
	if a.Count != nil {
		v, _ := strconv.Atoi(*a.Count)
		o.loopCount = int8(v)
	}
	if a.Value != nil {
		v, _ := strconv.ParseFloat(*a.Value, 32)
		o.volume = float32(v)
	}

	return
}

func (o *ObjectAudio) update(delta time.Duration) {
	o.elapsed += delta
	if o.elapsed >= 2*time.Second {
		t := o.GetTile()
		// Clean up listeners that don't exist
		for id := range o.listeners {
			if _, ok := t.GetMap().activeObjects[id]; !ok {
				delete(o.listeners, id)
			}
		}
		// Check for new listeners.
		for _, ao := range t.GetMap().activeObjects {
			id := ao.GetID()
			_, exists := o.listeners[id]

			switch obj := ao.(type) {
			case *ObjectCharacter:
				if obj.CanHear(obj.GetDistance(t.Y, t.X, t.Z)) {
					if !exists {
						obj.GetOwner().SendMusic(o.Audio(), o.SoundSet(), o.SoundIndex(), o.id, t.Y, t.X, t.Z, o.Volume(), o.LoopCount())
						o.listeners[id] = struct{}{}
					}
				} else {
					if exists {
						obj.GetOwner().StopMusic(o.id)
						delete(o.listeners, id)
					}
				}
			}
		}

		o.elapsed = 0
	}

	o.Object.update(delta)
}

// getType returns the Archetype type.
func (o *ObjectAudio) getType() data.ArchetypeType {
	return data.ArchetypeAudio
}

// Volume does an obvious thing.
func (o *ObjectAudio) Volume() float32 {
	return o.volume
}

// LoopCount returns the loop count of the object. -1 means loop, 0 means do nothing, 1 means play once, and so on.
func (o *ObjectAudio) LoopCount() int8 {
	return o.loopCount
}

// Audio returns the underlying audio id.
func (o *ObjectAudio) Audio() uint32 {
	return o.Archetype.AudioID
}

// SoundSet returns the underlying SoundSet id.
func (o *ObjectAudio) SoundSet() uint32 {
	return o.Archetype.SoundSetID
}

// SoundIndex returns the underlying SoundIndex.
func (o *ObjectAudio) SoundIndex() int8 {
	return o.Archetype.SoundIndex
}
