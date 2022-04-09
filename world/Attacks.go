package world

import (
	cdata "github.com/chimera-rpg/go-common/data"
)

// Attacks provides a mapping of AttackTypes to integer values. This is used for damage reduction and similar after damage rolls are made.
type Attacks map[cdata.AttackType]int
