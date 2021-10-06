package world

import (
	"errors"
	"time"

	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
	log "github.com/sirupsen/logrus"
)

// ObjectCharacter represents player characters.
type ObjectCharacter struct {
	Object
	//
	name                   string
	maxHp                  int
	level                  int
	race                   string
	count                  int
	value                  int
	mapUpdateTime          uint8 // Corresponds to the map's updateTime -- if they are out of sync then the player will sample its view space.
	resistances            data.AttackTypes
	attacktypes            data.AttackTypes
	attributes             data.Attributes
	skills                 []ObjectSkill
	equipment              []ObjectI // Equipment is all equipped inventory items.
	currentCommand         OwnerCommand
	currentCommandElapsed  time.Duration
	currentCommandDuration time.Duration
	// FIXME: Temporary code for testing a stamina system.
	stamina    time.Duration
	maxStamina time.Duration
}

// NewObjectCharacter creates a new ObjectCharacter from the given archetype.
func NewObjectCharacter(a *data.Archetype) (o *ObjectCharacter) {
	o = &ObjectCharacter{
		Object:     NewObject(a),
		maxStamina: time.Millisecond * 100, // TEMP.
	}

	// Create a new Owner AI if it is an NPC.
	if a.Type == cdata.ArchetypeNPC {
		owner := NewOwnerSimpleAI()
		owner.SetTarget(o)
	}
	// NOTE: We could/should probably have other AI types that can control multiple objects.

	//o.setArchetype(a)

	return
}

// NewObjectCharacterFromCharacter creates a new ObjectCharacter from the given character data.
func NewObjectCharacterFromCharacter(c *data.Character) (o *ObjectCharacter) {
	o = &ObjectCharacter{
		Object:     NewObject(&c.Archetype),
		name:       c.Name,
		maxStamina: time.Millisecond * 100, // TEMP.
	}
	return
}

func (o *ObjectCharacter) setArchetype(targetArch *data.Archetype) {
	// First inherit from another Archetype if ArchID is set.
	/*mutatedArch := data.NewArchetype()
	for targetArch != nil {
		if err := mergo.Merge(&mutatedArch, targetArch); err != nil {
			log.Fatal("o no")
		}
		targetArch = targetArch.InheritArch
	}

	o.name, _ = mutatedArch.Name.GetString()*/
}

func (o *ObjectCharacter) update(delta time.Duration) {
	doTilesBlock := func(targetTiles []*Tile) bool {
		isBlocked := false
		matter := o.GetArchetype().Matter
		for _, tT := range targetTiles {
			for _, tO := range tT.objects {
				switch tO := tO.(type) {
				case *ObjectBlock:
					// Check if the target object blocks our matter.
					if tO.blocking.Is(matter) {
						isBlocked = true
					}
				}
			}
		}
		return isBlocked
	}

	// Add a falling timer if we've moved and should fall.
	var s *StatusFalling
	if o.hasMoved && !o.HasStatus(s) {
		m := o.tile.gameMap
		if m != nil {
			_, fallingTiles, err := m.GetObjectPartTiles(o, -1, 0, 0, false)

			if !doTilesBlock(fallingTiles) && err == nil {
				o.AddStatus(&StatusFalling{})
			}
		}
	}

	o.stamina += delta
	if o.stamina > o.maxStamina {
		o.stamina = o.maxStamina
	}

	// Process as many commands as we can.
	isRunning := false
	for {
		// Always prioritize repeat commands.
		if cmd := o.GetOwner().RepeatCommand(); cmd != nil {
			cmd := cmd.(OwnerRepeatCommand)
			switch c := cmd.Command.(type) {
			case OwnerMoveCommand:
				if o.stamina >= 50*time.Millisecond {
					if _, err := o.GetTile().GetMap().MoveObject(o, c.Y, c.X, c.Z, false); err != nil {
						log.Warn(err)
					}
					o.stamina -= 50 * time.Millisecond
				}
				isRunning = true
			}
			break
		}
		if o.currentCommand == nil {
			if o.GetOwner() != nil {
				if o.GetOwner().HasCommands() {
					o.currentCommand = o.GetOwner().ShiftCommand()
					o.currentCommandDuration = time.Millisecond * 50
				} else {
					break
				}
			}
		}
		if o.stamina >= o.currentCommandDuration {
			switch c := o.currentCommand.(type) {
			case OwnerMoveCommand:
				if _, err := o.GetTile().GetMap().MoveObject(o, c.Y, c.X, c.Z, false); err != nil {
					log.Warn(err)
				}
			case OwnerStatusCommand:
				o.SetStatus(c.Status)
			}
			o.currentCommand = nil
			o.stamina -= o.currentCommandDuration
		} else {
			break
		}
	}
	if isRunning && !o.HasStatus(&StatusRunning{}) {
		o.AddStatus(&StatusRunning{})
	} else if !isRunning && o.HasStatus(&StatusRunning{}) {
		o.RemoveStatus(&StatusRunning{})
	}

	//
	o.Object.update(delta)
}

// AddStatus adds the given status to the character.
func (o *ObjectCharacter) AddStatus(s StatusI) {
	s.SetTarget(o)
	o.statuses = append(o.statuses, s)
	s.OnAdd()
	if o.GetOwner() != nil {
		o.GetOwner().SendStatus(s, true)
	}
}

// SetStatus sets the status.
func (o *ObjectCharacter) SetStatus(s StatusI) bool {
	switch s.(type) {
	case *StatusSqueeze:
		if o.HasStatus(&StatusCrouch{}) {
			o.GetOwner().SendMessage("You cannot squeeze while crouching.")
			return false
		}
		s2 := o.GetStatus(s)
		if s2 == nil {
			o.AddStatus(s)
			o.GetTile().GetMap().MoveObject(o, 0, 0, 0, false)
		} else {
			s2.(*StatusSqueeze).Remove = true
			o.GetTile().GetMap().MoveObject(o, 0, 0, 0, false)
		}

	case *StatusCrouch:
		if o.HasStatus(&StatusSqueeze{}) {
			o.GetOwner().SendMessage("You cannot crouch while squeezing.")
			return false
		}
		s2 := o.GetStatus(s)
		if s2 == nil {
			o.AddStatus(s)
			o.GetTile().GetMap().MoveObject(o, 0, 0, 0, false)
		} else {
			s2.(*StatusCrouch).Remove = true
			o.GetTile().GetMap().MoveObject(o, 0, 0, 0, false)
		}
	}
	return false
}

// ResolveEvent handles events that pertain to the character.
func (o *ObjectCharacter) ResolveEvent(e EventI) bool {
	// TODO: Send event messages to the owner.
	switch e := e.(type) {
	case EventFall:
		if o.GetOwner() != nil {
			o.GetOwner().SendMessage("You begin to fall...")
		}
		return true
	case EventFell:
		if o.GetOwner() != nil {
			o.GetOwner().SendMessage(e.String())
		}
		// TODO: If we're not invisible or very quiet, notify other creatures in a radius that we've cratered our legs.
		return true
		/*case EventSqueezing:
			if o.GetOwner() != nil {
				o.GetOwner().SendMessage("You are squeezing.")
			}
			return true
		case EventUnsqueeze:
			if o.GetOwner() != nil {
				o.GetOwner().SendMessage("You are no longer squeezing.")
			}
			return true*/
	}
	return false
}

// RollAttack rolls an attack with the given weapon.
func (o *ObjectCharacter) RollAttack(w *ObjectWeapon) (a Attacks) {
	//
	return a
}

func (o *ObjectCharacter) getType() cdata.ArchetypeType {
	return cdata.ArchetypePC
}

// EquipObject attempts to equip a given object. The object must be in the character's inventory and must be equippable (weapon, shield, or armor).
// This removes the item from the inventory and adds it to the equipment if the equipping was successful.
func (o *ObjectCharacter) EquipObject(ob ObjectI) error {
	// Ensure we are only equipping from our inventory.
	index := -1
	for i, v := range o.inventory {
		if v == o {
			index = i
			break
		}
	}
	if index == -1 {
		return errors.New("object does not exist in inventory")
	}
	// Ensure the item can be equipped.
	var err error
	switch obj := ob.(type) {
	case *ObjectWeapon:
		err = o.EquipWeapon(obj)
	case *ObjectShield:
		err = o.EquipShield(obj)
	case *ObjectArmor:
		err = o.EquipArmor(obj)
	default:
		return errors.New("object cannot be equipped")
	}
	if err == nil {
		o.equipment = append(o.equipment, o.inventory[index])
		o.inventory = append(o.inventory[:index], o.inventory[index+1:]...)
	}
	return err
}

// EquipArmor equips armor.
func (o *ObjectCharacter) EquipArmor(armor *ObjectArmor) error {
	return nil
}

// EquipShield equips a shield.
func (o *ObjectCharacter) EquipShield(armor *ObjectShield) error {
	return nil
}

// EquipWeapon equips a weapon.
func (o *ObjectCharacter) EquipWeapon(armor *ObjectWeapon) error {
	return nil
}

// Name returns the name of the character.
func (o *ObjectCharacter) Name() string {
	return o.name
}

// Stamina returns the object's stamina.
func (o *ObjectCharacter) Stamina() time.Duration {
	return o.stamina
}

// Stamina returns the object's max stamina.
func (o *ObjectCharacter) MaxStamina() time.Duration {
	return o.maxStamina
}
