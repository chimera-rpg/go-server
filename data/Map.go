package data

// Map is the source containing structure used to build a world.Map.
type Map struct {
	DataName    string
	Name        string
	Description string
	Lore        string
	Width       int
	Height      int
	Darkness    int
	ResetTime   int
	EastMap     string
	WestMap     string
	SouthMap    string
	NorthMap    string
	Tiles       [][][]Archetype
}
