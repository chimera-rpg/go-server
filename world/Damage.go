package world

import (
	"fmt"
	"strings"

	cdata "github.com/chimera-rpg/go-common/data"
)

type Damage struct {
	AttackType      cdata.AttackType
	BaseDamage      float64
	AttributeDamage float64
	Styles          map[cdata.AttackStyle]float64
	StyleTotal      float64
}

type Damages []Damage

func (ds *Damages) Total() (total float64) {
	for _, d := range *ds {
		styleDamages := 0.0
		attributeDamages := 0.0
		for _, s := range d.Styles {
			styleDamages += d.BaseDamage * s
			attributeDamages += d.AttributeDamage * s / d.StyleTotal
		}
		total += styleDamages
		total += attributeDamages
	}
	return total
}

func (ds Damages) String() string {
	var styleStrings []string
	var total float64
	var totalAttributes float64
	for _, d := range ds {
		styleDamages := 0.0
		attributeDamages := 0.0
		for k, s := range d.Styles {
			styleStrings = append(styleStrings, fmt.Sprintf("%.1f %s", d.BaseDamage*s, cdata.AttackStyleToStringMap[k]))
			styleDamages += d.BaseDamage * s
			attributeDamages += d.AttributeDamage * s / d.StyleTotal
		}
		total += styleDamages
		total += attributeDamages
		totalAttributes += attributeDamages
	}
	return fmt.Sprintf("%.1f (%s) [%.1f Attr]", total, strings.Join(styleStrings, ", "), totalAttributes)
}

func (ds *Damages) Clone() (ds2 Damages) {
	for _, d := range *ds {
		d2 := Damage{
			AttackType:      d.AttackType,
			BaseDamage:      d.BaseDamage,
			AttributeDamage: d.AttributeDamage,
			Styles:          make(map[cdata.AttackStyle]float64),
			StyleTotal:      d.StyleTotal,
		}
		for k, v := range d.Styles {
			d2.Styles[k] = v
		}
		ds2 = append(ds2, d2)
	}

	return ds2
}
