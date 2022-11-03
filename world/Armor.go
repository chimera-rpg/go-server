package world

import (
	"fmt"
	"strings"

	"github.com/chimera-rpg/go-server/data"
)

/*
 */
type Armor struct {
	ArmorType data.AttackType              // Physical, etc.
	Styles    map[data.AttackStyle]float64 // Impact, etc.
}

type Armors []Armor

func (as Armors) String() (s string) {
	for _, a := range as {
		s += data.AttackTypeToStringMap[a.ArmorType] + "("
		var styleStrings []string
		for k, st := range a.Styles {
			styleStrings = append(styleStrings, fmt.Sprintf("%s: %1.f%%", data.AttackStyleToStringMap[k], st*100))
		}
		s += strings.Join(styleStrings, ",") + ") "
	}
	return
}

func (as *Armors) Reduce(ds *Damages) {
	for _, d := range *ds {
		// Get our matching armor.
		var armor *Armor
		for i, a := range *as {
			if a.ArmorType == d.AttackType {
				armor = &(*as)[i]
				break
			}
		}
		if armor == nil {
			continue
		}

		// Reduce the damage style according to our skill and styles.
		for attackStyle, damageStyle := range d.Styles {
			if armorStyle, ok := armor.Styles[attackStyle]; ok {
				d.Styles[attackStyle] = damageStyle - armorStyle
				if d.Styles[attackStyle] < 0 {
					d.Styles[attackStyle] = 0
				}
			}
		}
	}
}

// Merge merges 2 armors, additively.
func (as *Armors) Merge(as2 Armors) {
	for _, a2 := range as2 {
		for _, a := range *as {
			if a.ArmorType == a2.ArmorType {
				for armorStyle2, armorValue2 := range a2.Styles {
					if _, ok := a.Styles[armorStyle2]; ok {
						a.Styles[armorStyle2] += armorValue2
					} else {
						a.Styles[armorStyle2] = armorValue2
					}
				}
			}
		}
	}
}

func (as Armors) Clone() (as2 Armors) {
	for _, a := range as {
		a2 := Armor{
			ArmorType: a.ArmorType,
			Styles:    make(map[data.AttackStyle]float64),
		}
		for k, v := range a.Styles {
			a2.Styles[k] = v
		}
		as2 = append(as2, a2)
	}

	return as2
}
