package data

// AnimationsConfig is the configuration for animations. It is loaded from the archetypes root by the server and sent to clients on connection.
type AnimationsConfig struct {
	TileWidth  uint8 `yaml:"TileWidth,omitempty"`
	TileHeight uint8 `yaml:"TileHeight,omitempty"`
	YStep      struct {
		X int8 `yaml:"X,omitempty"`
		Y int8 `yaml:"Y,omitempty"`
	} `yaml:"YStep,omitempty"`
	Adjustments map[ArchetypeType]struct {
		X int8 `yaml:"X,omitempty"`
		Y int8 `yaml:"Y,omitempty"`
	} `yaml:"Adjustments,omitempty"`
}
