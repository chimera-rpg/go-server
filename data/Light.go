package data

type Light struct {
	Brightness float32 `yaml:"Brightness,omitempty"` // Brightness represents... something
	Red        float32 `yaml:"Red,omitempty"`        // Multiplier for red
	Green      float32 `yaml:"Green,omitempty"`      // Multiplier for green
	Blue       float32 `yaml:"Blue,omitempty"`       // Mulitplier for blue
	Intensity  float32 `yaml:"Intensity,omitempty"`  // Intensity represents how far the light travels.
}

// Add adds other's values to ourself.
func (l *Light) Add(other *Light) {
	l.Brightness += other.Brightness
	l.Red += other.Red
	l.Green += other.Green
	l.Intensity += other.Intensity
}
