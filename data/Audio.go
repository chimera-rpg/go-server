package data

// AudioPre
type AudioPre struct {
	SoundSets map[string][]AudioSoundPre `yaml:"SoundSets"`
}

// AudioSoundPre
type AudioSoundPre struct {
	File string `yaml:"File"` // During post-parsing this is used to acquire and set the SoundID.
	Text string `yaml:"Text"`
}

// Audio
type Audio struct {
	SoundSets map[StringID][]AudioSound
}

// AudioSound
type AudioSound struct {
	SoundID StringID
	Text    string
}
