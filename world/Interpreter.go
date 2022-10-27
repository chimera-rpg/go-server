package world

import (
	"github.com/chimera-rpg/go-server/data"
	"github.com/cosmos72/gomacro/fast"
)

// SetupInterpreterTypes sets up an interpreter with our common world types.
func SetupInterpreterTypes(interp *fast.Interp) {
	interp.ImportPackage("fmt", "fmt")
	interp.ImportPackage("time", "time")

	// Interfaces
	interp.DeclType(interp.TypeOf((*OwnerI)(nil)).Elem())
	interp.DeclType(interp.TypeOf((*ObjectI)(nil)).Elem())
	interp.DeclType(interp.TypeOf((*EventI)(nil)).Elem())
	interp.DeclType(interp.TypeOf((*ActionI)(nil)).Elem())

	// Concrete Types
	interp.DeclType(interp.TypeOf(EventAdvance{}))
	interp.DeclType(interp.TypeOf(EventAttacked{}))
	interp.DeclType(interp.TypeOf(EventAttacking{}))
	interp.DeclType(interp.TypeOf(EventAttack{}))
	interp.DeclType(interp.TypeOf(EventBirth{}))
	interp.DeclType(interp.TypeOf(EventDestroy{}))
	interp.DeclType(interp.TypeOf(EventFall{}))
	interp.DeclType(interp.TypeOf(EventFell{}))
	interp.DeclType(interp.TypeOf(EventExit{}))

	interp.DeclType(interp.TypeOf(OwnerPlayer{}))
	interp.DeclType(interp.TypeOf(OwnerSimpleAI{}))

	interp.DeclType(interp.TypeOf(Object{}))
	interp.DeclType(interp.TypeOf(ObjectEquippable{}))
	interp.DeclType(interp.TypeOf(ObjectAudio{}))
	interp.DeclType(interp.TypeOf(ObjectBlock{}))
	interp.DeclType(interp.TypeOf(ObjectCharacter{}))
	interp.DeclType(interp.TypeOf(ObjectExit{}))
	interp.DeclType(interp.TypeOf(ObjectFlora{}))
	interp.DeclType(interp.TypeOf(ObjectFood{}))
	interp.DeclType(interp.TypeOf(ObjectGeneric{}))
	interp.DeclType(interp.TypeOf(ObjectItem{}))
	interp.DeclType(interp.TypeOf(ObjectSkill{}))
	interp.DeclType(interp.TypeOf(ObjectTile{}))

	interp.DeclType(interp.TypeOf(ActionMove{}))
	interp.DeclType(interp.TypeOf(ActionAttack{}))
	interp.DeclType(interp.TypeOf(ActionSpawn{}))
	interp.DeclType(interp.TypeOf(ActionStatus{}))

	interp.DeclType(interp.TypeOf(data.Duration{}))

	interp.DeclType(interp.TypeOf(Map{}))
	interp.DeclType(interp.TypeOf(Tile{}))
}
