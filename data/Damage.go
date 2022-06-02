package data

import (
	cdata "github.com/chimera-rpg/go-common/data"
)

type Damage struct {
	Value          int                    `json:"Value" yaml:"Value,omitempty"`
	AttributeBonus DamageAttributeBonuses `json:"AttributeBonus" ts_type:"{[key:number]: {[key:number]: number}}" yaml:"AttributeBonus,omitempty"`
}

func (d *Damage) Add(other *Damage) {
	d.Value += other.Value
	for k, v := range other.AttributeBonus {
		if _, ok := d.AttributeBonus[k]; !ok {
			d.AttributeBonus[k] = v
		} else {
			for k2, v2 := range v {
				if _, ok := d.AttributeBonus[k][k2]; !ok {
					d.AttributeBonus[k][k2] = v2
				} else {
					d.AttributeBonus[k][k2] += v2
				}
			}
		}
	}
}

type DamageAttributeBonuses map[cdata.AttackType]AttributeTypes

type AttributeTypes map[AttributeType]float64

// UnmarshalYAML unmarshals, converting attack type strings into AttackTypes.
func (d *DamageAttributeBonuses) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*d = make(DamageAttributeBonuses)
	var value map[string]map[string]float64
	if err := unmarshal(&value); err != nil {
		return err
	}
	for k, v := range value {
		if ak, ok := cdata.StringToAttackTypeMap[k]; ok {
			(*d)[ak] = make(map[AttributeType]float64)
			for k2, v2 := range v {
				if sk, ok := StringToAttributeTypeMap[k2]; ok {
					(*d)[ak][sk] = v2
				}
			}
		}
	}
	return nil
}

// MarshalYAML marshals, converting AttackTypes into strings.
func (d DamageAttributeBonuses) MarshalYAML() (interface{}, error) {
	r := make(map[string]map[string]float64)
	for k, v := range d {
		if sk, ok := cdata.AttackTypeToStringMap[k]; ok {
			r[sk] = make(map[string]float64)
			for k2, v2 := range v {
				if sk2, ok := AttributeTypeToStringMap[k2]; ok {
					r[sk][sk2] = v2
				}
			}
		}
	}
	return r, nil
}
