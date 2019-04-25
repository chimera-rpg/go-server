package world

/*
The following system of Ability Score rules is known as the Xibabla SR and is copyright MindFire Games, LLC.
*/

// Kinetic represents the physical presence of a character.
type Kinetic struct {
	// Power is the muscularity of the kinetic ability class. It is vital for unarmed and melee combat, as it is used in activities such as swinging a sword, wrestling, or drawing a bow.
	Power int
	// Finesse pertains to the character's quickness, agility, and ability to dance out of the way of danger.
	Finesse int
}

// Techne represents the ability of the character to do fine tasks.
type Techne struct {
	// Precision is used for ranged attacks, fine crafting, and accurate movements with one's hands.
	Precision int
	// Reason is used along-side Precision for disarming and crafting traps, along with many other crafting or magic-based abilities. Additionally, it is used to see through illusions and otherwise.
	Reason int
}

// Stoic represents physicality and the ability to push past suffering.
type Stoic struct {
	// Endurance is used in saves against poison and the ability to press on while injured. It is also used alongside Power to determine hit points.
	Endurance int
	// Willpower is the ability to push beyond what you can normally stand. High willpower allows you to bypass the penalties of poison, exhaustion, sickness, and injury. It also provides the ability to resist mental domination.
	Willpower int
}

// Social is used for interactions between people and otherwise.
type Social struct {
	// Empathy is used for interpersonal situations, such as brokering a deal, calming an ally, or befriending an NPC. It also has an influence over interacting with animals and sensing a person's intentions.
	Empathy int
	// Guile is used for diplomacy, performing, bluffing, and similar.
	Guile int
}

// Perception is your ability to see and be a part of the world.
type Perception struct {
	// Insight is used for actively scanning for objects or things. Influences the ability to see invisible creatures and detecting traps.
	Insight int
	// Sense is used for passively sensing the world around you, such as when a creature makes a noise far away, or when someone is stealthed.
	Sense int
}

// AbilityScores is the collection of abilities and their values.
type AbilityScores struct {
	Kinetic    Kinetic
	Stoic      Stoic
	Techne     Techne
	Perception Perception
	Social     Social
}

// CalculateHealthBonus calculates the health bonus provided by the passed AbilityScores.
func (as AbilityScores) CalculateHealthBonus() (hp int) {
	if as.Stoic.Endurance < 0 {
		hp = hp + as.Stoic.Endurance/2
	} else if as.Stoic.Endurance > 0 {
		hp = hp + as.Stoic.Endurance
	}
	if as.Kinetic.Power < 0 {
		hp = hp + as.Kinetic.Power/4
	} else if as.Kinetic.Power > 0 {
		hp = hp + (as.Kinetic.Power / 2)
	}
	return
}
