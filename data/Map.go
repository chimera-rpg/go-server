package data

// Map is the source containing structure used to build a world.Map.
type Map struct {
	MapID        StringID          `json:"-" yaml:"-"`
	DataName     string            `json:"-" yaml:"-"`
	Filepath     string            `json:"-" yaml:"-"`
	Name         string            `json:"Name" yaml:"Name"`
	Description  string            `json:"Description" yaml:"Description"`
	Lore         string            `json:"Lore" yaml:"Lore"`
	Depth        int               `json:"Depth" yaml:"Depth"`
	Width        int               `json:"Width" yaml:"Width"`
	Height       int               `json:"Height" yaml:"Height"`
	AmbientRed   uint8             `json:"AmbientRed" yaml:"AmbientRed"`
	AmbientGreen uint8             `json:"AmbientGreen" yaml:"AmbientGreen"`
	AmbientBlue  uint8             `json:"AmbientBlue" yaml:"AmbientBlue"`
	Outdoor      bool              `json:"Outdoor" yaml:"Outdoor"`
	OutdoorRed   uint8             `json:"OutdoorRed" yaml:"OutdoorRed"`
	OutdoorGreen uint8             `json:"OutdoorGreen" yaml:"OutdoorGreen"`
	OutdoorBlue  uint8             `json:"OutdoorBlue" yaml:"OutdoorBlue"`
	ResetTime    int               `json:"ResetTime" yaml:"ResetTime"`
	Y            int               `json:"Y" yaml:"Y"`
	X            int               `json:"X" yaml:"X"`
	Z            int               `json:"Z" yaml:"Z"`
	Haven        bool              `json:"Haven" yaml:"Haven"`
	Tiles        [][][][]Archetype `json:"Tiles" yaml:"Tiles"`
	Script       string            `json:"Script" yaml:"Script"` // Script is stored as full code, as each map data instance holds its own complete interpreter.
}
