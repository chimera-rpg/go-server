package world

import (
	"github.com/chimera-rpg/go-server/data"
)

// Attacks provides a mapping of data.AttackTypes to integer values. This is used for damage reduction and similar after damage rolls are made.
type Attacks map[data.AttackType]int
