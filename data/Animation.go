package data

// Animation represents a collection of data that is used for managing
// Object animation.
type Animation struct {
	AnimID uint32
	Faces  map[uint32][]AnimationFrame
}

// AnimationFrame represents an individual frame of an animation.
type AnimationFrame struct {
	ImageID   uint32
	FrameTime int
}
