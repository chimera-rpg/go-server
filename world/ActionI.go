package world

import "time"

// ActionI represents actions such as moving or attacking.
type ActionI interface {
	ChannelTime() time.Duration  // Time preceding the action's result.
	RecoveryTime() time.Duration // Time succeeding the action's result.
	Channel(bool)
	Channeled() bool
}
