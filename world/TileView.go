package world

// TileView represents an Owner's view of a particular tile.
type TileView struct {
	unset    bool
	visible  bool
	modTime  uint16   // corresponds to the modTime of whatever tile this is supposed to reference.
	knownIDs []uint32 // List of IDs last known in this tile.
}
