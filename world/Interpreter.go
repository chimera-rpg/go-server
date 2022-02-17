package world

import (
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

	// Concrete Types
	interp.DeclType(interp.TypeOf(EventAdvance{}))
	interp.DeclType(interp.TypeOf(EventBirth{}))
	interp.DeclType(interp.TypeOf(EventDestroy{}))
	interp.DeclType(interp.TypeOf(EventFall{}))
	interp.DeclType(interp.TypeOf(EventFell{}))
	interp.DeclType(interp.TypeOf(EventExit{}))

	interp.DeclType(interp.TypeOf(OwnerPlayer{}))
	interp.DeclType(interp.TypeOf(OwnerSimpleAI{}))

	interp.DeclType(interp.TypeOf(Object{}))
	interp.DeclType(interp.TypeOf(ObjectArmor{}))
	interp.DeclType(interp.TypeOf(ObjectAudio{}))
	interp.DeclType(interp.TypeOf(ObjectBlock{}))
	interp.DeclType(interp.TypeOf(ObjectCharacter{}))
	interp.DeclType(interp.TypeOf(ObjectExit{}))
	interp.DeclType(interp.TypeOf(ObjectFlora{}))
	interp.DeclType(interp.TypeOf(ObjectFood{}))
	interp.DeclType(interp.TypeOf(ObjectGeneric{}))
	interp.DeclType(interp.TypeOf(ObjectItem{}))
	interp.DeclType(interp.TypeOf(ObjectShield{}))
	interp.DeclType(interp.TypeOf(ObjectSkill{}))
	interp.DeclType(interp.TypeOf(ObjectTile{}))
	interp.DeclType(interp.TypeOf(ObjectWeapon{}))

	interp.DeclType(interp.TypeOf(Map{}))
	interp.DeclType(interp.TypeOf(Tile{}))
}
