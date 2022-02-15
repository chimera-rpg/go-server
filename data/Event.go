package data

import (
	"fmt"
	"math/rand"

	cdata "github.com/chimera-rpg/go-common/data"
)

// SpawnEventType represents the spawn event rule to use.
type SpawnEventType uint8

// UnmarhsalYAML converts a SpawnEventType string to its numerical value.
func (stype *SpawnEventType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var value string
	if err := unmarshal(&value); err != nil {
		return err
	}
	if v, ok := StringToSpawnEventType[value]; ok {
		*stype = v
		return nil
	}
	*stype = RandomSpawn

	return fmt.Errorf("unknown SpawnEventType '%s'", value)
}

// Our spawn rules.
const (
	RandomSpawn SpawnEventType = iota
	ExclusiveWeightedSpawn
)

// StringToSpawnEventType does as it says.
var StringToSpawnEventType = map[string]SpawnEventType{
	"Random":             RandomSpawn,
	"Exclusive Weighted": ExclusiveWeightedSpawn,
}

// SpawnEventTypeToString does as it says.
var SpawnEventTypeToString = map[SpawnEventType]string{
	RandomSpawn:            "Random",
	ExclusiveWeightedSpawn: "Exclusive Weighted",
}

// TimeRange just defines a range between two times.
type IntRange struct {
	Min int   `yaml:"Min,omitempty"`
	Max int   `yaml:"Max,omitempty"`
	Not []int `yaml:"Not,omitempty"`
}

// Random returns a count between min and max, inclusive.
func (t IntRange) Random() int {
	if t.Min == t.Max {
		return t.Max
	}
	// Iffy code to reroll if Not has values.
	if len(t.Not) > 0 {
		var r int
		for done := false; !done; {
			r = rand.Intn(t.Max-t.Min+1) + t.Min
			match := 0
			for _, i := range t.Not {
				if r == i {
					match++
				}
			}
			if match == 0 {
				done = true
			}
		}
		return r
	}
	return rand.Intn(t.Max-t.Min+1) + t.Min
}

type SpawnConditions struct {
	Matter   *cdata.MatterType `yaml:"Matter,omitempty"`
	Blocking *cdata.MatterType `yaml:"Blocking,omitempty"`
}

type SpawnPlacement struct {
	Overlap bool            `yaml:"Overlap,omitempty"`
	Surface SpawnConditions `yaml:"Surface,omitempty"`
	Air     SpawnConditions `yaml:"Air,omitempty"`
	X       IntRange        `yaml:"X,omitempty"`
	Y       IntRange        `yaml:"Y,omitempty"`
	Z       IntRange        `yaml:"Z,omitempty"`
}

// SpawnArchetype is a container for pairing an Archetype with a Chance value.
type SpawnArchetype struct {
	Chance    float32        `yaml:"Chance,omitempty"`
	Archetype *Archetype     `yaml:"Archetype,omitempty"`
	Count     IntRange       `yaml:"Count,omitempty"`
	Retry     int            `yaml:"Retry,omitempty"` // Retry count if spawning fails.
	Placement SpawnPlacement `yaml:"Placement,omitempty"`
}

type XYZOffset struct {
	X IntRange `yaml:"X,omitempty"`
	Y IntRange `yaml:"Y,omitempty"`
	Z IntRange `yaml:"Z,omitempty"`
}

// SpawnEventResponse is an event response that causes zero or more items to spawn.
type SpawnEventResponse struct {
	Type  SpawnEventType   `yaml:"Type,omitempty"`
	Items []SpawnArchetype `yaml:"Items,omitempty"`
}

// TriggerEventResponse causes another event to be triggered.
type TriggerEventResponse struct {
	Event string `yaml:"Event,omitempty"`
}

// EventType is a type representing our events.
type EventType uint32

// Events is a structure for our various potentially handled event responses.
type Events struct {
	Birth   *EventResponses `yaml:"Birth,omitempty"`
	Death   *EventResponses `yaml:"Death,omitempty"`
	Hit     *EventResponses `yaml:"Hit,omitempty"`
	Advance *EventResponses `yaml:"Advance,omitempty"`
}

// EventResponses are the valid responses that can be given for an event.
type EventResponses struct {
	Spawn   *SpawnEventResponse   `yaml:"Spawn,omitempty"`
	Replace *[]SpawnArchetype     `yaml:"Replace,omitempty"`
	Trigger *TriggerEventResponse `yaml:"Trigger,omitempty"`
	Script  *ScriptEventResponse  `yaml:"Script,omitempty"`
}
