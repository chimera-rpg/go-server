package data

type AttackStyle uint32

const (
	NoAttackStyle AttackStyle = iota
	// Impact represents physical attack types of a bludgeoning nature.
	Impact
	// Pierce represents physical attack types of a sharp, thrusting nature.
	Pierce
	// Edged represents physical attack types of an edged nature, such as swords.
	Edged
	// Flame represents fire.
	Flame
	// Frost represents cold.
	Frost
	// Lightning represents Zeus's mighty power.
	Lightning
	// Corrosive represents acid.
	Corrosive
	// Force represents mere force.
	Force
	// Heal does healing.
	Heal
	// Harm does harming.
	Harm
)

var StringToAttackStyleMap = map[string]AttackStyle{
	"Impact": Impact,
	"Pierce": Pierce,
	"Edged":  Edged,

	"Flame":     Flame,
	"Frost":     Frost,
	"Lightning": Lightning,
	"Corrosive": Corrosive,
	"Force":     Force,

	"Heal": Heal,
	"Harm": Harm,
}

var AttackStyleToStringMap = map[AttackStyle]string{
	Impact: "Impact",
	Pierce: "Pierce",
	Edged:  "Edged",

	Flame:     "Flame",
	Frost:     "Frost",
	Lightning: "Lightning",
	Corrosive: "Corrosive",
	Force:     "Force",

	Heal: "Heal",
	Harm: "Harm",
}
