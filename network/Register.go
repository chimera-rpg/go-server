package network

import (
	"encoding/gob"
)

// RegisterCommands registers our various Command structures with their gob names.
func RegisterCommands() {
	gob.RegisterName("H", CommandHandshake{})
	gob.RegisterName("F", CommandFeatures{})
	gob.RegisterName("B", CommandBasic{})
	gob.RegisterName("M", CommandMap{})
	gob.RegisterName("L", CommandLogin{})
	gob.RegisterName("R", CommandRejoin{})
	gob.RegisterName("C", CommandQueryCharacters{})
	gob.RegisterName("CC", CommandCharacter{})
	gob.RegisterName("C+", CommandCreateCharacter{})
	gob.RegisterName("C_", CommandSelectCharacter{})
	gob.RegisterName("Cg", CommandQueryGenera{})
	gob.RegisterName("Cs", CommandQuerySpecies{})
	gob.RegisterName("Cv", CommandQueryVariety{})
	gob.RegisterName("Cc", CommandQueryCulture{})
	gob.RegisterName("Cl", CommandQueryLegacy{})
	gob.RegisterName("Ct", CommandQueryTraining{})
	gob.RegisterName("A", CommandAnimation{})
	gob.RegisterName("G", CommandGraphics{})
	gob.RegisterName("T", CommandTile{})
	gob.RegisterName("Tt", CommandTiles{})
	gob.RegisterName("Tl", CommandTileLight{})
	gob.RegisterName("Ts", CommandTileSky{})
	gob.RegisterName("O", CommandObject{})
	gob.RegisterName("Oc", CommandObjectPayloadCreate{})
	gob.RegisterName("Od", CommandObjectPayloadDelete{})
	gob.RegisterName("Oa", CommandObjectPayloadAnimate{})
	gob.RegisterName("Ov", CommandObjectPayloadViewTarget{})
	gob.RegisterName("Oi", CommandObjectPayloadInfo{})
	gob.RegisterName("OC", CommandObjectPayloadContainer{})
	gob.RegisterName("c", CommandCmd{})
	gob.RegisterName("cl", CommandClearCmd{})
	gob.RegisterName("e", CommandExtCmd{})
	gob.RegisterName("r", CommandRepeatCmd{})
	gob.RegisterName("m", CommandMessage{})
	gob.RegisterName("s", CommandStatus{})
	gob.RegisterName("t", CommandStamina{})
	gob.RegisterName("I", CommandInspect{})
	gob.RegisterName("Vp", CommandViewport{})
	gob.RegisterName("S", CommandSound{})
	gob.RegisterName("a", CommandAudio{})
	gob.RegisterName("n", CommandNoise{})
	gob.RegisterName("Mu", CommandMusic{})
	gob.RegisterName("At", CommandAttack{})
	gob.RegisterName("D", CommandDamage{})
	gob.RegisterName("In", CommandInteract{})
}
