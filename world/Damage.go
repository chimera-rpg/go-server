package world

import "github.com/chimera-rpg/go-server/data"

type Damage struct {
	AttackType   data.AttackType
	AttackStyles map[data.AttackStyle]float64
	BaseDamage   float64
	Competency   float64
	Skill        float64
}
