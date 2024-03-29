package world

import (
	"time"

	"github.com/chimera-rpg/go-server/data"
	"github.com/chimera-rpg/go-server/network"
)

// OwnerI represents the general interface that should be used
// for controlling and managing autonomous Object(s). It is used for
// Players and will eventually be used for NPCs.
type OwnerI interface {
	GetTarget() ObjectI
	SetTarget(ObjectI)
	SetMap(*Map)
	GetMap() *Map
	Update(delta time.Duration) error
	OnMapUpdate(delta time.Duration) error
	OnObjectDelete(ID) error
	SetViewSize(h, w, d int)
	GetViewSize() (h, w, d int)
	//
	HasAttitude(ID) bool
	GetAttitude(ID, bool) data.Attitude
	//
	SendCommand(network.Command) error
	SendMessage(string)
	SendStatus(StatusI, bool)
	SendSound(audioID ID, soundID ID, objectID ID, y, x, z int, volume float32)
	SendMusic(audioID, soundID ID, soundIndex int8, objectID ID, y, x, z int, volume float32, loopCount int8)
	StopMusic(objectID ID)
	//
	RepeatCommand() OwnerCommand
	HasCommands() bool
	PushCommand(c OwnerCommand)
	ShiftCommand() OwnerCommand
	ClearCommands()
	Wizard() bool
	//
	ForgetObject(ID)
}
