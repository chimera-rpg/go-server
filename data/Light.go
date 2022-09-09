package data

type Light struct {
	Red      uint8   `yaml:"Red,omitempty"`      // Multiplier for red
	Green    uint8   `yaml:"Green,omitempty"`    // Multiplier for green
	Blue     uint8   `yaml:"Blue,omitempty"`     // Multiplier for blue
	Distance float64 `yaml:"Distance,omitempty"` // Distance the light travels.
	Falloff  float64 `yaml:"Falloff,omitempty"`  // How fast the light falls off, relative to its distance, in terms of 0..1 of distance.
}

// Add adds other's values to ourself.
func (l *Light) Add(other *Light) {
	l.Red += other.Red
	l.Green += other.Green
	l.Blue += other.Blue
	l.Distance += other.Distance
	l.Falloff += other.Falloff
}
