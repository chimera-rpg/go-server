package data

// AnimationPre represents a collection of data that is used for managing Object animation.
type AnimationPre struct {
	//	AnimID StringID
	Faces map[string][]AnimationFramePre `yaml:"Faces"`
}

// AnimationFramePre represents an individual frame of an animation.
type AnimationFramePre struct {
	Image string `yaml:"Image"` // During post-parsing this is used to acquire and set the ImageID.
	Time  int    `yaml:"Time"`
	X     int8   `yaml:"X"`
	Y     int8   `yaml:"Y"`
}

// Animation represents a collection of data that is used for managing Object animation.
type Animation struct {
	//	AnimID StringID
	Faces map[StringID][]AnimationFrame
}

// AnimationFrame represents an individual frame of an animation.
type AnimationFrame struct {
	ImageID StringID
	Time    int
	X, Y    int8 // Allow X and Y offset adjustments
}
