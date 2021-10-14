package world

import (
	"strconv"

	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
)

// ObjectAudio represents audio for in-map sound and music playback.
type ObjectAudio struct {
	Object
	loopCount int
	volume    int
}

// NewObjectAudio creates a sound and music playback object from the given archetype.
func NewObjectAudio(a *data.Archetype) (o *ObjectAudio) {
	o = &ObjectAudio{
		Object: NewObject(a),
	}
	// Use count for loops and value for volume.
	if a.Count != nil {
		o.loopCount, _ = strconv.Atoi(*a.Count)
	}
	if a.Value != nil {
		o.volume, _ = strconv.Atoi(*a.Value)
	}

	return
}

// getType returns the Archetype type.
func (o *ObjectAudio) getType() cdata.ArchetypeType {
	return cdata.ArchetypeAudio
}

// Volume does an obvious thing.
func (o *ObjectAudio) Volume() int {
	return o.volume
}

// LoopCount returns the loop count of the object. -1 means loop, 0 means do nothing, 1 means play once, and so on.
func (o *ObjectAudio) LoopCount() int {
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
