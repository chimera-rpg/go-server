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
	maxHp         int
	race          string
	count         int
	value         int
	mapUpdateTime uint8 // Corresponds to the map's updateTime -- if they are out of sync then the player will sample its view space.
	// Fields that are pointers to the underlying archetype. In the case of NPCs, the archetypes should correspond to the instance of their on-map Archetype or the result of an archetype spawn. In the case of PCs, the archetype corresponds to the one embedded in their player data file.
	name         *string
	level        *int
	resistances  *data.AttackTypes
	attacktypes  *data.AttackTypes
	attributes   *data.AttributeSets
	competencies *map[data.CompetencyType]data.Competency
	skills       []ObjectSkill
	//
	equipment             []ObjectI // Equipment is all equipped inventory items.
	currentActionDuration time.Duration
	// FIXME: Temporary code for testing a stamina system.
	stamina                int // Stamina is a pool that recharges and is consumed by actions.
	maxStamina             int
	currentAction          ActionI
	speedPenaltyMultiplier int // The speed penalty multiplier, for status penalties to action duration costs.
	//
	shouldRecalculate bool
	speed             int
	health            int
}

// NewObjectCharacter creates a new ObjectCharacter from the given archetype.
func NewObjectCharacter(a *data.Archetype) (o *ObjectCharacter) {
	o = &ObjectCharacter{
		Object:                 NewObject(a),
		speedPenaltyMultiplier: 1,
	}
	*o.name = *a.Name
	*o.level = a.Level
	*o.resistances = a.Resistances
	*o.attacktypes = a.AttackTypes
	*o.attributes = a.Attributes
	*o.competencies = a.Competencies
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
func NewObjectCharacterFromCharacter(c *data.Character, completeArchetype *data.Archetype) (o *ObjectCharacter) {
	o = &ObjectCharacter{
		Object:                 NewObject(completeArchetype),
		name:                   &c.Name,
		level:                  &c.Archetype.Level,
		resistances:            &c.Archetype.Resistances,
		attacktypes:            &c.Archetype.AttackTypes,
		attributes:             &c.Archetype.Attributes,
		competencies:           &c.Archetype.Competencies,
		speedPenaltyMultiplier: 1,
	}
	o.AltArchetype = &c.Archetype
	//o.maxStamina = time.Duration(o.CalculateStamina())
	o.maxStamina = o.CalculateStamina()
	o.Recalculate()
	// TODO: Move elsewhere.
	/*for statusID, statusMap := range c.SaveInfo.Statuses {
		if statusID == int(cdata.CrouchingStatus) {
			s := &StatusCrouch{}
			s.Deserialize(statusMap)
		} else if statusID == int(cdata.SqueezingStatus) {
			s := &StatusSqueeze{}
			s.Deserialize(statusMap)
		} else if statusID == int(cdata.FallingStatus) {
			s := &StatusFalling{}
			s.Deserialize(statusMap)
		}
	}*/
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
	if o.shouldRecalculate {
		o.Recalculate()
		o.shouldRecalculate = false
	}

	doTilesBlock := func(targetTiles []*Tile) bool {
		isBlocked := false
		matter := o.GetArchetype().Matter
		for _, tT := range targetTiles {
			for _, tO := range tT.objectParts {
				switch tO := tO.(type) {
				case *ObjectTile:
					if tO.blocking.Is(matter) {
						isBlocked = true
						break
					}
				case *ObjectBlock:
					if tO.blocking.Is(matter) {
						isBlocked = true
						break
					}
				}
			}
			if isBlocked {
				break
			}
		}
		return isBlocked
	}

	// Add a falling timer if we've moved and should fall.
	if o.hasMoved && !o.HasStatus(StatusFallingRef) {
		m := o.tile.gameMap
		if m != nil {
			_, fallingTiles, err := m.GetObjectPartTiles(o, -1, 0, 0, false)

			if !o.HasStatus(StatusFlyingRef) && !doTilesBlock(fallingTiles) && err == nil {
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
				duration := calcDuration(100*time.Millisecond, 20*time.Millisecond, time.Duration(o.speed)*time.Millisecond)
				o.currentAction = NewActionMove(c.Y, c.X, c.Z, duration, true)
				o.currentActionDuration = 0 // TODO: Add remainder from last operation if possible.
			}
		} else if o.GetOwner() != nil && o.GetOwner().HasCommands() {
			cmd := o.GetOwner().ShiftCommand()
			switch c := cmd.(type) {
			case OwnerMoveCommand:
				duration := calcDuration(100*time.Millisecond, 20*time.Millisecond, time.Duration(o.speed)*time.Millisecond)
				o.currentAction = NewActionMove(c.Y, c.X, c.Z, duration, false)
				o.currentActionDuration = 0 // TODO: Add remainder from last operation if possible.
			case OwnerStatusCommand:
				duration := calcDuration(200*time.Millisecond, 50*time.Millisecond, time.Duration(o.speed)*time.Millisecond)
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
			t := o.GetTile()
			_, w, d := o.GetDimensions()
			audioID := t.GetMap().world.data.Strings.Acquire("thump")
			soundID := t.GetMap().world.data.Strings.Acquire("default")
			t.GetMap().EmitSound(audioID, soundID, t.y-1, t.x+w/2, t.z+d/2, 0.25)
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
	return *o.name
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

// Recalculate caches all the stats of the player.
func (o *ObjectCharacter) Recalculate() {
	o.speed = o.CalculateSpeed()
	o.health = o.CalculateHealth()
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
	result := 10 // Baseline 10.

	// Add any bonuses from our ancestry.
	for _, a := range o.Archetype.ArchPointers {
		result += int(a.Attributes.Physical.GetSpeedBonus())
	}
	// Add from our own archetype.
	result += int(o.AltArchetype.Attributes.Physical.GetSpeedBonus())

	return result
}

// CalculateHealth calculates the maximum health a character has.
func (o *ObjectCharacter) CalculateHealth() int {
	result := 0

	// Add any bonuses from our ancestry.
	for _, a := range o.Archetype.ArchPointers {
		result += int(a.Attributes.Physical.GetHealthBonus())
	}
	// Add from our own archetype.
	result += int(o.AltArchetype.Attributes.Physical.GetHealthBonus())

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

// CanHear returns whether or not the character can hear an object at a given distance units away.
func (o *Object) CanHear(distance float64) bool {
	// TODO: Use Sense + Focus(minor)
	return true // FIXME
}

// HandleSound processes a sound at a given coordinate to see if it should be heard.
func (o *ObjectCharacter) HandleSound(audioID, soundID ID, y, x, z int, volume float32) {
	d := o.GetDistance(y, x, z)
	if d < 100*float64(volume) { // FIXME: Use a Sense + Focus(minor) derived value.
		o.GetOwner().SendSound(audioID, soundID, 0, y, x, z, volume)
		// TODO: EventSound?
	}
}

// HandleObjectSound processes a sound from a given object to see if it should be heard.
func (o *ObjectCharacter) HandleObjectSound(audioID, soundID ID, o2 ObjectI, volume float32) {
	t2 := o.GetTile()
	d := o.GetDistance(t2.y, t2.x, t2.z)
	if d < 100*float64(volume) { // FIXME: Use a Sense + Focus(minor) derived value.
		o.GetOwner().SendSound(audioID, soundID, o2.GetID(), 0, 0, 0, volume)
		// TODO: EventSound?
	}
}
