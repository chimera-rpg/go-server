package data

// Map is the source containing structure used to build a world.Map.
type Map struct {
	MapID       StringID          `yaml:"-"`
	DataName    string            `yaml:"-"`
	Name        string            `yaml:"Name"`
	Description string            `yaml:"Description"`
	Lore        string            `yaml:"Lore"`
	Depth       int               `yaml:"Depth"`
	Width       int               `yaml:"Width"`
	Height      int               `yaml:"Height"`
	Darkness    int               `yaml:"Darkness"`
	ResetTime   int               `yaml:"ResetTime"`
	Tiles       [][][][]Archetype `yaml:"Tiles"`
}
