package world

import (
	"fmt"
	"strings"

	"github.com/chimera-rpg/go-server/data"
)

type Damage struct {
	AttackType      data.AttackType
	BaseDamage      float64
	AttributeDamage float64
	Styles          map[data.AttackStyle]float64
}

type Damages []Damage

func (ds *Damages) Total() (total float64) {
	for _, d := range *ds {
		styleDamages := 0.0
		for _, s := range d.Styles {
			styleDamages += d.BaseDamage * s
		}
		total += styleDamages
	}
	return total
}

func (ds *Damages) String() string {
	var styleStrings []string
	var total float64
	for _, d := range *ds {
		styleDamages := 0.0
		for k, s := range d.Styles {
			styleStrings = append(styleStrings, fmt.Sprintf("%.1f %s", d.BaseDamage*s, data.AttackStyleToStringMap[k]))
			styleDamages += d.BaseDamage * s
		}
		total += styleDamages
	}
	return fmt.Sprintf("%.1f (%s)", total, strings.Join(styleStrings, ", "))
}

func (ds *Damages) Clone() (ds2 Damages) {
	for _, d := range *ds {
		d2 := Damage{
			AttackType:      d.AttackType,
			BaseDamage:      d.BaseDamage,
			AttributeDamage: d.AttributeDamage,
			Styles:          make(map[data.AttackStyle]float64),
		}
		for k, v := range d.Styles {
			d2.Styles[k] = v
		}
		ds2 = append(ds2, d2)
	}

	return ds2
}
