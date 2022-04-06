package world

import "github.com/chimera-rpg/go-server/data"

type Damage struct {
	AttackType      data.AttackType
	StyleDamages    map[data.AttackStyle]float64
	AttributeDamage float64
}
