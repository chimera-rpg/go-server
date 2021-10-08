package world

import (
	"errors"
	"fmt"
	"time"

	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
	log "github.com/sirupsen/logrus"
)

// ObjectCharacter represents player characters.
type ObjectCharacter struct {
	Object
	//
	name                  string
	maxHp                 int
	level                 int
	race                  string
	count                 int
	value                 int
	mapUpdateTime         uint8 // Corresponds to the map's updateTime -- if they are out of sync then the player will sample its view space.
	resistances           data.AttackTypes
	attacktypes           data.AttackTypes
	attributes            data.AttributeSets
	skills                []ObjectSkill
	equipment             []ObjectI // Equipment is all equipped inventory items.
	currentActionDuration time.Duration
	// FIXME: Temporary code for testing a stamina system.
	stamina                int // Stamina is a pool that recharges and is consumed by actions.
	maxStamina             int
	currentAction          ActionI
	speedPenaltyMultiplier int // The speed penalty multiplier, for status penalties to action duration costs.
}

// NewObjectCharacter creates a new ObjectCharacter from the given archetype.
func NewObjectCharacter(a *data.Archetype) (o *ObjectCharacter) {
	o = &ObjectCharacter{
		Object:                 NewObject(a),
		speedPenaltyMultiplier: 1,
	}
	o.maxStamina = o.CalculateStamina()

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
		Object:                 NewObject(&c.Archetype),
		name:                   c.Name,
		speedPenaltyMultiplier: 1,
	}
	//o.maxStamina = time.Duration(o.CalculateStamina())
	o.maxStamina = o.CalculateStamina()
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

	// Process as many commands as we can.
	o.currentActionDuration += delta / time.Duration(o.speedPenaltyMultiplier)
	for {
		// Process our current action first.
		if o.currentAction != nil {
			if !o.currentAction.Channeled() && o.currentActionDuration >= o.currentAction.ChannelTime() {
				switch a := o.currentAction.(type) {
				case *ActionMove:
					if _, err := o.GetTile().GetMap().MoveObject(o, a.y, a.x, a.z, false); err != nil {
						log.Warn(err)
					}
				case *ActionStatus:
					o.SetStatus(a.status)
				}
				o.currentAction.Channel(true)
			}
			if o.currentActionDuration < o.currentAction.ChannelTime()+o.currentAction.RecoveryTime() {
				break
			}
			o.currentAction = nil
		}
		calcDuration := func(base time.Duration, min time.Duration, reduction time.Duration) time.Duration {
			d := base - reduction
			if d < min {
				d = min
			}
			return d
		}
		// Always prioritize repeat commands.
		if cmd := o.GetOwner().RepeatCommand(); cmd != nil {
			cmd := cmd.(OwnerRepeatCommand)
			switch c := cmd.Command.(type) {
			case OwnerMoveCommand:
				// Cap movement duration cost to a minimum of 20 millisecond
				duration := calcDuration(100*time.Millisecond, 20*time.Millisecond, time.Duration(o.CalculateSpeed())*time.Millisecond)
				o.currentAction = NewActionMove(c.Y, c.X, c.Z, duration, true)
				o.currentActionDuration = 0 // TODO: Add remainder from last operation if possible.
			}
		} else if o.GetOwner() != nil && o.GetOwner().HasCommands() {
			cmd := o.GetOwner().ShiftCommand()
			switch c := cmd.(type) {
			case OwnerMoveCommand:
				duration := calcDuration(100*time.Millisecond, 20*time.Millisecond, time.Duration(o.CalculateSpeed())*time.Millisecond)
				o.currentAction = NewActionMove(c.Y, c.X, c.Z, duration, false)
				o.currentActionDuration = 0 // TODO: Add remainder from last operation if possible.
			case OwnerStatusCommand:
				duration := calcDuration(200*time.Millisecond, 50*time.Millisecond, time.Duration(o.CalculateSpeed())*time.Millisecond)
				o.currentAction = NewActionStatus(c.Status, duration)
				o.currentActionDuration = 0 // TODO: Add remainder from last operation if possible.
			}
		}
		if o.currentAction == nil {
			break
		}
	}
	isRunning := false

	if o.currentAction != nil {
		switch a := o.currentAction.(type) {
		case *ActionMove:
			if a.running {
				isRunning = true
			}
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
	switch s.(type) {
	case *StatusSqueeze:
		o.speedPenaltyMultiplier++
	case *StatusCrouch:
		o.speedPenaltyMultiplier += 2
	}
}

// RemoveStatus removes the given status type from the character.
func (o *ObjectCharacter) RemoveStatus(s StatusI) StatusI {
	if s = o.Object.RemoveStatus(s); s != nil {
		switch s.(type) {
		case *StatusSqueeze:
			o.speedPenaltyMultiplier--
		case *StatusCrouch:
			o.speedPenaltyMultiplier -= 2
		}
	}
	return s
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
func (o *ObjectCharacter) Stamina() int {
	return o.stamina
}

// MaxStamina returns the object's max stamina.
func (o *ObjectCharacter) MaxStamina() int {
	return o.maxStamina
}

// RestoreStamina restores a calculated amount of stamina on turn start.
func (o *ObjectCharacter) RestoreStamina() {
	if o.stamina < o.maxStamina {
		o.stamina += o.CalculateStamina()
		if o.stamina > o.maxStamina {
			o.stamina = o.maxStamina
		}
	}
}

// CalculateStamina calculates the maximum stamina based upon our attributes.
func (o *ObjectCharacter) CalculateStamina() int {
	result := 1 // Baseline 1, so the player can always somewhat move.

	p, _ := o.GetAttributeValue(data.PhysicalAttributes, data.Prowess)
	m, _ := o.GetAttributeValue(data.PhysicalAttributes, data.Might)

	result += int(p) + int(m)/2

	return result
}

// CalculateSpeed calculates the speed a character has.
func (o *ObjectCharacter) CalculateSpeed() int {
	result := 10

	h, _ := o.GetAttributeValue(data.PhysicalAttributes, data.Haste)
	r, _ := o.GetAttributeValue(data.PhysicalAttributes, data.Reaction)

	result += int(h)*5 + int(r)/4*5

	return result
}

// CalculateHealth calculates the maximum health a character has.
func (o *ObjectCharacter) CalculateHealth() int {
	result := 0

	p, _ := o.GetAttributeValue(data.PhysicalAttributes, data.Prowess)
	m, _ := o.GetAttributeValue(data.PhysicalAttributes, data.Might)

	result += int(p) * 8
	result += int(m) * 2

	return result
}

// GetAttributeValue gets the calculated value for a given attribute.
func (o *ObjectCharacter) GetAttributeValue(set data.AttributesType, a data.AttributeType) (int, error) {
	var value int
	var archSet, objSet *data.Attributes

	switch set {
	case data.PhysicalAttributes:
		archSet = &o.Archetype.Attributes.Physical
		objSet = &o.attributes.Physical
	case data.ArcaneAttributes:
		archSet = &o.Archetype.Attributes.Arcane
		objSet = &o.attributes.Arcane
	case data.SpiritAttributes:
		archSet = &o.Archetype.Attributes.Spirit
		objSet = &o.attributes.Spirit
	default:
		return 0, fmt.Errorf("no attribute set matching %d", set)
	}

	// TODO: Calculate bonuses from items, statuses, and beyond!
	switch a {
	case data.Might:
		value += int(archSet.Might)
		value += int(objSet.Might)
	case data.Prowess:
		value += int(archSet.Prowess)
		value += int(objSet.Prowess)
	case data.Focus:
		value += int(archSet.Focus)
		value += int(objSet.Focus)
	case data.Sense:
		value += int(archSet.Sense)
		value += int(objSet.Sense)
	case data.Haste:
		value += int(archSet.Haste)
		value += int(objSet.Haste)
	case data.Reaction:
		value += int(archSet.Reaction)
		value += int(objSet.Reaction)
	default:
		return 0, fmt.Errorf("no attribute matching %d", a)
	}
	return value, nil
}
