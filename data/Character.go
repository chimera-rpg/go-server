package data

// Character represents a player character. It is largely just
// an Archetype...? Maybe it should just be that? What about
// race and class? Hmm.
// Perhaps Character should be: Archetype, RaceArchetype, ClassArchetype,
// all of which are built upon one another?
type Character struct {
	Name          string    `yaml:"Name"`
	Archetype     Archetype `yaml:"Archetype"`
	RaceArchetype *Archetype
	Race          string `yaml:"Race"`
	Class         string `yaml:"Class"`
	//	ClassArchetype *Archetype
}
