package data

// Do you got the 'tude, man?

import (
	"fmt"
)

// Attitude represents the variable levels of hostility or friendliness.
type Attitude uint8

// These attitudes correspond to the given levels of hostility or friendliness of a character.
const (
	NoAttitude      Attitude = 0
	SlavishAttitude          = 1 << iota
	AlliedAttitude
	FriendlyAttitude
	NeutralAttitude
	UnfriendlyAttitude
	HostileAttitude
	LoathingAttitude
)

// StringToAttitudeMap is a map of strings to Attitude types.
var StringToAttitudeMap = map[string]Attitude{
	"None":       NoAttitude,
	"Slavish":    SlavishAttitude,
	"Allied":     AlliedAttitude,
	"Friendly":   FriendlyAttitude,
	"Neutral":    NeutralAttitude,
	"Unfriendly": UnfriendlyAttitude,
	"Hostile":    HostileAttitude,
	"Loathing":   LoathingAttitude,
}

// AttitudeToStringMap is a map of Attitude types to strings.
var AttitudeToStringMap = map[Attitude]string{
	NoAttitude:         "None",
	SlavishAttitude:    "Slavish",
	AlliedAttitude:     "Allied",
	FriendlyAttitude:   "Friendly",
	NeutralAttitude:    "Neutral",
	UnfriendlyAttitude: "Unfriendly",
	HostileAttitude:    "Hostile",
	LoathingAttitude:   "Loathing",
}

// UnmarshalYAML unmarshals an Attitude from a string.
func (a *Attitude) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var value string
	if err := unmarshal(&value); err != nil {
		return err
	}
	if v, ok := StringToAttitudeMap[value]; ok {
		*a = v
		return nil
	}
	*a = NoAttitude
	return fmt.Errorf("Unknown Attitude '%s'", value)
}

// MarshalYAML marshals an Attitude into a string.
func (a Attitude) MarshalYAML() (interface{}, error) {
	if v, ok := AttitudeToStringMap[a]; ok {
		return v, nil
	}
	return "None", nil
}

// Attitudes contain families and factions that are considered for attitude relation.
type Attitudes struct {
	Families map[string]FamilyAttitudes `yaml:"Families,omitempty"`
	Factions map[string]Attitudes       `yaml:"Factions,omitempty"`
}

// FamilyAttitudes contain genera that are considered for attitude relation.
type FamilyAttitudes struct {
	Genera map[string]GeneraAttitudes `yaml:"Genera,omitempty"`
}

// GeneraAttitudes contain species that are considered for attitude relation.
type GeneraAttitudes struct {
	Species map[string]Attitude `yaml:"Species,omitempty"`
}
