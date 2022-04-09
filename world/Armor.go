package world

import "github.com/chimera-rpg/go-server/data"

/*

 */
type Armor struct {
	ArmorType data.AttackType // Physical, etc.
	//BaseArmor  float64                      // 0 thru 1 (0% to 100%)
	ArmorSkill float64                      // (skill * competence)
	Styles     map[data.AttackStyle]float64 // Impact, etc.
}

type Armors []Armor

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
				d.Styles[attackStyle] = damageStyle - (armorStyle * armor.ArmorSkill)
				if d.Styles[attackStyle] < 0 {
					d.Styles[attackStyle] = 0
				}
			}
		}
	}
}
