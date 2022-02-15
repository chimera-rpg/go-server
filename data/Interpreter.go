package data

import (
	"github.com/cosmos72/gomacro/fast"
)

// Interpreter is the interpreter instance used by all archetype scripts. This is more fully defined in world.
var Interpreter *fast.Interp = fast.New()
