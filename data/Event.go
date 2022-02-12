package data

import "fmt"

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

// SpawnArchetype is a container for pairing an Archetype with a Chance value.
type SpawnArchetype struct {
	Chance    float32    `yaml:"Chance,omitempty"`
	Archetype *Archetype `yaml:"Archetype,omitempty"`
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
	// Script *ScriptEventResponse `yaml:"Script,omitempty"`
}
