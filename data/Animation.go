package data

// Animation represents a collection of data that is used for managing
// Object animation.
type Animation struct {
	AnimId uint32
	//Faces map[string][]AnimationFrame
	Faces map[uint32][]AnimationFrame
}

type AnimationFrame struct {
	//FilePath string
	ImageId   uint32
	FrameTime int
}
