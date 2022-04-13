package world

import (
	"strings"

	"github.com/chimera-rpg/go-server/data"
)

type Owner struct {
	commandQueue  []OwnerCommand
	repeatCommand OwnerCommand
	wizard        bool
	attitudes     map[ID]data.Attitude
	currentMap    *Map
	target        *ObjectCharacter
}

func (owner *Owner) HasCommands() bool {
	return len(owner.commandQueue) > 0
}

func (owner *Owner) PushCommand(c OwnerCommand) {
	owner.commandQueue = append(owner.commandQueue, c)
}

func (owner *Owner) ShiftCommand() OwnerCommand {
	c := owner.commandQueue[0]
	owner.commandQueue = owner.commandQueue[1:]
	return c
}

func (owner *Owner) ClearCommands() {
	owner.commandQueue = make([]OwnerCommand, 0)
	owner.repeatCommand = nil
}

func (owner *Owner) RepeatCommand() OwnerCommand {
	return owner.repeatCommand
}

func (owner *Owner) Wizard() bool {
	return owner.wizard
}

func (owner *Owner) ForgetObject(oID ID) {
}

// HasAttitude returns if the owner has an attitude towards the given object.
func (owner *Owner) HasAttitude(oID ID) bool {
	_, ok := owner.attitudes[oID]
	return ok
}

// GetAttitude returns the attitude the owner has the a given object. If no attitude exists, one is calculated based upon the target's attitude (if it has one).
func (owner *Owner) GetAttitude(oID ID, recurse bool) data.Attitude {
	if attitude, ok := owner.attitudes[oID]; ok {
		return attitude
	}
	target := owner.GetMap().world.GetObject(oID)
	if target == nil {
		delete(owner.attitudes, oID)
	} else {
		attitude := data.NoAttitude
		if ownerArchetype := owner.target.GetArchetype(); ownerArchetype != nil {
			if targetArchetype := target.GetArchetype(); targetArchetype != nil {
				// First check against our own faction attitudes
				// TODO: These faction comparisons should have more detail in terms of combinatory operations, such as allowing FactionA, !FactionB, etc. -- it should compare all faction relations, then determine the final attitude from that.
				for faction, value := range ownerArchetype.Attitudes.Factions {
					if faction[0] == '!' {
						s := strings.TrimPrefix(faction, "!")
						has := false
						for _, otherFaction := range targetArchetype.Factions {
							if otherFaction == s {
								has = true
								break
							}
						}
						if !has {
							attitude = value
						}
					} else {
						for _, otherFaction := range targetArchetype.Factions {
							if faction == otherFaction {
								attitude = value
								break
							}
						}
					}
					if attitude != data.NoAttitude {
						break
					}
				}
				// Second check against species -> genera.
				if attitude == data.NoAttitude {
					if g, ok := ownerArchetype.Attitudes.Genera[targetArchetype.Genera]; ok {
						attitude = g.Attitude
						if s := g.Species[targetArchetype.Species]; ok {
							attitude = s
						}
					}
				}
				// Third check against legacy.
				if attitude == data.NoAttitude {
					if l, ok := ownerArchetype.Attitudes.Legacies[targetArchetype.Legacy]; ok {
						attitude = l
					}
				}
			}
		}
		// If we have no attitude at this point, try to default from the other owner's attitudes.
		if attitude == data.NoAttitude && recurse {
			if otherOwner := target.GetOwner(); otherOwner != nil {
				if otherOwner.HasAttitude(owner.target.id) {
					attitude = otherOwner.GetAttitude(owner.target.id, false)
				}
			}
		}
		owner.attitudes[oID] = attitude
		return attitude
	}

	return data.NoAttitude
}

// GetMap gets the currentMap of the owner.
func (owner *Owner) GetMap() *Map {
	return owner.currentMap
}
