package data

// MapArchetype represents an Archetype that, when Map creation is desired, will be merged with the Archetype it references in "Arch" and output an Object based upon this combination.
/*type MapArchetype struct {
  Arch string
  archetype Archetype
}*/

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
