package data

// AudioPre
type AudioPre struct {
	SoundSets map[string][]AudioSoundPre `json:"SoundSets" yaml:"SoundSets"`
}

// AudioSoundPre
type AudioSoundPre struct {
	File string `json:"File" yaml:"File"` // During post-parsing this is used to acquire and set the SoundID.
	Text string `json:"Text" yaml:"Text"`
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
