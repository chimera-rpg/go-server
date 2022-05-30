package data

// Map is the source containing structure used to build a world.Map.
type Map struct {
	MapID       StringID          `json:"-" yaml:"-"`
	DataName    string            `json:"-" yaml:"-"`
	Name        string            `json:"Name" yaml:"Name"`
	Description string            `json:"Description" yaml:"Description"`
	Lore        string            `json:"Lore" yaml:"Lore"`
	Depth       int               `json:"Depth" yaml:"Depth"`
	Width       int               `json:"Width" yaml:"Width"`
	Height      int               `json:"Height" yaml:"Height"`
	Darkness    int               `json:"Darkness" yaml:"Darkness"`
	ResetTime   int               `json:"ResetTime" yaml:"ResetTime"`
	Y           int               `json:"Y" yaml:"Y"`
	X           int               `json:"X" yaml:"X"`
	Z           int               `json:"Z" yaml:"Z"`
	Tiles       [][][][]Archetype `json:"Tiles" yaml:"Tiles"`
	Script      string            `json:"Script" yaml:"Script"` // Script is stored as full code, as each map data instance holds its own complete interpreter.
}
