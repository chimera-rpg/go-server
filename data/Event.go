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
	Matter   *cdata.MatterType `json:"Matter" yaml:"Matter,omitempty"`
	Blocking *cdata.MatterType `json:"Blocking" yaml:"Blocking,omitempty"`
}

type SpawnPlacement struct {
	Overlap bool            `json:"Overlap" yaml:"Overlap,omitempty"`
	Surface SpawnConditions `json:"Surface" yaml:"Surface,omitempty"`
	Air     SpawnConditions `json:"Air" yaml:"Air,omitempty"`
	X       IntRange        `json:"X" yaml:"X,omitempty"`
	Y       IntRange        `json:"Y" yaml:"Y,omitempty"`
	Z       IntRange        `json:"Z" yaml:"Z,omitempty"`
}

// SpawnArchetype is a container for pairing an Archetype with a Chance value.
type SpawnArchetype struct {
	Chance    float32        `json:"Chance" yaml:"Chance,omitempty"`
	Archetype *Archetype     `json:"Archetype" yaml:"Archetype,omitempty"`
	Count     IntRange       `json:"Count" yaml:"Count,omitempty"`
	Retry     int            `json:"Retry" yaml:"Retry,omitempty"` // Retry count if spawning fails.
	Placement SpawnPlacement `json:"Placement" yaml:"Placement,omitempty"`
}

type XYZOffset struct {
	X IntRange `json:"X" yaml:"X,omitempty"`
	Y IntRange `json:"Y" yaml:"Y,omitempty"`
	Z IntRange `json:"Z" yaml:"Z,omitempty"`
}

// SpawnEventResponse is an event response that causes zero or more items to spawn.
type SpawnEventResponse struct {
	Type  SpawnEventType   `json:"Type" yaml:"Type,omitempty"`
	Items []SpawnArchetype `json:"Items" yaml:"Items,omitempty"`
}

// TriggerEventResponse causes another event to be triggered.
type TriggerEventResponse struct {
	Event string `json:"Event" yaml:"Event,omitempty"`
}

// EventType is a type representing our events.
type EventType uint32

// Events is a structure for our various potentially handled event responses.
type Events struct {
	Birth     *EventResponses `json:"Birth" yaml:"Birth,omitempty"`
	Death     *EventResponses `json:"Death" yaml:"Death,omitempty"`
	Hit       *EventResponses `json:"Hit" yaml:"Hit,omitempty"`
	Attacking *EventResponses `json:"Attacking" yaml:"Attacking,omitempty"`
	Attacked  *EventResponses `json:"Attacked" yaml:"Attacked,omitempty"`
	Attack    *EventResponses `json:"Attack" yaml:"Attack,omitempty"`
	Advance   *EventResponses `json:"Advance" yaml:"Advance,omitempty"`
	Exit      *EventResponses `json:"Exit" yaml:"Exit,omitempty"`
}

// EventResponses are the valid responses that can be given for an event.
type EventResponses struct {
	Spawn   *SpawnEventResponse   `json:"Spawn" yaml:"Spawn,omitempty"`
	Replace *[]SpawnArchetype     `json:"Replace" yaml:"Replace,omitempty"`
	Trigger *TriggerEventResponse `json:"Trigger" yaml:"Trigger,omitempty"`
	Script  *ScriptEventResponse  `json:"Script" yaml:"Script,omitempty"`
}
