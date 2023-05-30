package world

import (
	"time"

	"github.com/chimera-rpg/go-server/data"
)

// ObjectI is the basic interface for Object access.
type ObjectI interface {
	GetID() ID
	GetOwner() OwnerI
	SetID(ID)
	SetOwner(OwnerI)
	SetMoved(bool)
	GetTile() *Tile
	SetTile(*Tile)
	setArchetype(*data.Archetype)
	GetArchetype() *data.Archetype
	ReplaceArchetype(a *data.Archetype)
	GetSaveableArchetype() data.Archetype
	//
	SetContainer(ObjectI)
	GetContainer() ObjectI
	//
	update(time.Duration)
	getType() data.ArchetypeType
	AddStatus(StatusI)
	RemoveStatus(StatusI) StatusI
	HasStatus(StatusI) bool
	SetStatus(StatusI) bool
	GetStatus(StatusI) StatusI
	ResolveEvent(EventI) bool
	Blocks(data.MatterType) bool
	Matter() data.MatterType
	SetName(string)
	Name() string
	GetDimensions() (h, w, d int)
	GetDistance(y, x, z int) float64
	ShootRay(y, x, z float64, f func(tile *Tile) bool) (tiles []*Tile)
	//
	GetMapID() ID
	InSameMap(ObjectI) bool
	//
	Stamina() int
	MaxStamina() int
	//
	RestoreStamina()
	//
	Attackable() bool
	//
	Updates() bool
	Timers() *[]Timer
	//
	Resistances() Armors
	//
	GetAttitude(ObjectI) data.Attitude
	//
	GetMundaneInfo(near bool) data.ObjectInfo
}
