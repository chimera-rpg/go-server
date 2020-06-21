package world

// OwnerI represents the general interface that should be used
// for controlling and managing autonomous Object(s). It is used for
// Players and will eventually be used for NPCs.
type OwnerI interface {
	GetTarget() ObjectI
	SetTarget(ObjectI)
	GetCommandChannel() chan OwnerCommand
	SetMap(*Map)
	GetMap() *Map
	Update(delta int64) error
	OnMapUpdate(delta int64) error
	OnObjectDelete(ID) error
	CreateView()
	CheckView()
}
