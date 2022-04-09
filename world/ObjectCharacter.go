package world

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-common/network"
	"github.com/chimera-rpg/go-server/data"
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
	resistances  *cdata.AttackTypes
	attacktypes  *cdata.AttackTypes
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
	damages []Damages
	armor   Armors
	dodge   float64
	//
	shouldRecalculate bool
	speed             int
	health            int
	reach             int
}

// NewObjectCharacter creates a new ObjectCharacter from the given archetype.
func NewObjectCharacter(a *data.Archetype) (o *ObjectCharacter) {
	o = &ObjectCharacter{
		Object:                 NewObject(a),
		speedPenaltyMultiplier: 1,
		reach:                  1,
	}
	*o.name = *a.Name
	*o.level = a.Level
	*o.resistances = a.Resistances
	*o.attacktypes = a.AttackTypes
	*o.attributes = a.Attributes
	*o.competencies = a.Competencies
	o.reach = int(a.Reach)
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
		reach:                  int(c.Archetype.Reach),
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

	// Add a falling timer if we've moved and should fall.
	if o.hasMoved {
		m := o.tile.gameMap
		if m != nil {
			_, tiles, err := m.GetObjectPartTiles(o, -1, 0, 0, false)
			// Check if we should be swimming.
			if err == nil {
				inLiquid := IsInLiquid(tiles)

				if inLiquid && !o.HasStatus(StatusSwimmingRef) {
					o.AddStatus(&StatusSwimming{})
				} else if !inLiquid && o.HasStatus(StatusSwimmingRef) {
					o.RemoveStatus(StatusSwimmingRef)
				}
				if !o.HasStatus(StatusFloatingRef) && !o.HasStatus(StatusSwimmingRef) && !o.HasStatus(StatusFlyingRef) && !o.HasStatus(StatusFallingRef) && !DoTilesBlock(o, tiles) {
					o.AddStatus(&StatusFalling{})
				}
			}
		}
	}

	// Adust action duration cooldown
	o.currentActionDuration += delta / time.Duration(o.speedPenaltyMultiplier)
	// Process our current action.
	if o.currentAction != nil {
		if !o.currentAction.Channeled() && o.currentActionDuration >= o.currentAction.ChannelTime() {
			o.currentAction.SetObject(o)
			o.currentAction.SetReady(true)
			o.tile.gameMap.QueueAction(o.currentAction)
			o.currentAction.Channel(true)
		}
		if o.currentActionDuration >= o.currentAction.ChannelTime()+o.currentAction.RecoveryTime() {
			o.currentAction = nil
		}
	}

	// Find a new action if we have any pending commands.
	if o.currentAction == nil {
		calcDuration := func(base time.Duration, min time.Duration, reduction time.Duration) time.Duration {
			d := base - reduction
			if d < min {
				d = min
			}
			return d
		}
		// FIXME: MOVE THIS
		buildAttackAction := func(c OwnerAttackCommand) *ActionAttack {
			base := 500 * time.Millisecond
			duration := calcDuration(base, 20*time.Millisecond, time.Duration(o.speed)*time.Millisecond)
			var y, x, z int
			if c.Y != 0 || c.X != 0 || c.Z != 0 {
				y = c.Y
				x = c.X
				z = c.Z
			} else if c.Target == 0 {
				h, w, d := o.GetDimensions()
				if c.Direction == network.North {
					z = -o.reach
				} else if c.Direction == network.South {
					z = o.reach + d
				} else if c.Direction == network.East {
					x = o.reach + w
				} else if c.Direction == network.West {
					x = -o.reach
				} else if c.Direction == network.Up {
					y = o.reach + h
				} else if c.Direction == network.Down {
					y = -o.reach
				}
				y = o.tile.Y + y
				x = o.tile.X + x
				z = o.tile.Z + z
			}
			return NewActionAttack(y, x, z, c.Target, duration)
		}

		// Always prioritize repeat commands.
		if cmd := o.GetOwner().RepeatCommand(); cmd != nil {
			cmd := cmd.(OwnerRepeatCommand)
			switch c := cmd.Command.(type) {
			case OwnerMoveCommand:
				base := 100 * time.Millisecond
				// If we're swimming, double the movement time.
				if o.HasStatus(StatusSwimmingRef) {
					// TODO: Adjust via a swimming speed.
					base *= 2
				} else if o.HasStatus(StatusFlyingRef) {
					// TODO: Adjust via a flying speed.
				} else if o.HasStatus(StatusFloatingRef) {
					// TODO: Also adjust via a floating speed.
					base *= 4 // For now, quadruple if we're floating.
				}
				// Cap movement duration cost to a minimum of 20 millisecond
				duration := calcDuration(base, 20*time.Millisecond, time.Duration(o.speed)*time.Millisecond)
				o.currentAction = NewActionMove(c.Y, c.X, c.Z, duration, true)
				o.currentActionDuration = 0 // TODO: Add remainder from last operation if possible.
			case OwnerAttackCommand:
				o.currentAction = buildAttackAction(c)
				o.currentActionDuration = 0 // TODO: Add remainder from last operation if possible.
			}
		} else if o.GetOwner() != nil && o.GetOwner().HasCommands() {
			cmd := o.GetOwner().ShiftCommand()
			switch c := cmd.(type) {
			case OwnerMoveCommand:
				base := 100 * time.Millisecond
				// If we're swimming, double the movement time.
				if o.HasStatus(StatusSwimmingRef) {
					// TODO: Adjust via a swimming speed.
					base *= 2
				} else if o.HasStatus(StatusFlyingRef) {
					// TODO: Adjust via a flying speed.
				} else if o.HasStatus(StatusFloatingRef) {
					// TODO: Also adjust via a floating speed.
					base *= 4 // For now, quadruple if we're floating.
				}
				// Cap movement duration cost to a minimum of 20 millisecond
				duration := calcDuration(base, 20*time.Millisecond, time.Duration(o.speed)*time.Millisecond)
				o.currentAction = NewActionMove(c.Y, c.X, c.Z, duration, false)
				o.currentActionDuration = 0 // TODO: Add remainder from last operation if possible.
			case OwnerAttackCommand:
				o.currentAction = buildAttackAction(c)
				o.currentActionDuration = 0 // TODO: Add remainder from last operation if possible.
			case OwnerStatusCommand:
				duration := calcDuration(200*time.Millisecond, 50*time.Millisecond, time.Duration(o.speed)*time.Millisecond)
				o.currentAction = NewActionStatus(c.Status, duration)
				o.currentActionDuration = 0 // TODO: Add remainder from last operation if possible.
			}
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
	case *StatusSwimming:
		t := o.GetTile()
		h, w, d := o.GetDimensions()
		audioID := t.GetMap().world.data.Strings.Acquire("water")
		soundID := t.GetMap().world.data.Strings.Acquire("enter")
		t.GetMap().EmitSound(audioID, soundID, t.Y+h-h/3, t.X+w/2, t.Z+d/2, 0.25)
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
		case *StatusSwimming:
			t := o.GetTile()
			h, w, d := o.GetDimensions()
			audioID := t.GetMap().world.data.Strings.Acquire("water")
			soundID := t.GetMap().world.data.Strings.Acquire("leave")
			t.GetMap().EmitSound(audioID, soundID, t.Y+h/2, t.X+w/2, t.Z+d/2, 0.25)
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
			h, w, d := o.GetDimensions()
			var audioID, soundID uint32
			if e.matter.Is(cdata.LiquidMatter) {
				audioID = t.GetMap().world.data.Strings.Acquire("water")
				soundID = t.GetMap().world.data.Strings.Acquire("sploosh")
				t.GetMap().EmitSound(audioID, soundID, t.Y+h-h/3, t.X+w/2, t.Z+d/2, 1.0)
			} else {
				audioID = t.GetMap().world.data.Strings.Acquire("thump")
				soundID = t.GetMap().world.data.Strings.Acquire("default")
				t.GetMap().EmitSound(audioID, soundID, t.Y-1, t.X+w/2, t.Z+d/2, 0.25)
			}
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
	case *EventAttack:
		var damageStrings []string
		for _, d := range e.Damages {
			damageStrings = append(damageStrings, d.String())
		}
		o.GetOwner().SendMessage(fmt.Sprintf("You attack for %s", strings.Join(damageStrings, ", ")))
		for _, ds := range e.Damages {
			dr := ds.Result()
			for _, d := range dr {
				o.GetOwner().SendCommand(network.CommandDamage{
					Target:          e.Target.GetID(),
					Type:            d.AttackType,
					AttributeDamage: d.AttributeDamage,
					StyleDamage:     d.Styles,
				})
			}
		}
	}
	// Resolve normal events.
	o.Object.ResolveEvent(e)
	return false
}

func (o *ObjectCharacter) Attack(o2 ObjectI) bool {
	//t2 := o2.GetTile()

	// FIXME: We should do a distance check, but it should be from the ideal facing edge.
	//distance := o.GetDistance(t2.Y, t2.X, t2.Z)
	//if float64(o.reach) >= distance {
	// FIXME: o.Matter() should be the weapon instead, as well as any spells/abilities of the character!
	if !o.Matter().Is(o2.Matter()) {
		fmt.Println("Our matter does not effect their matter")
		return false
	}

	// Clone our precalculated normal damages so they can be modified.
	var damages []Damages
	for _, d := range o.damages {
		damages = append(damages, d.Clone())
	}

	// Send the attacking event to the attacker.
	e1 := &EventAttacking{
		Target:  o2,
		Damages: damages,
	}
	o.ResolveEvent(e1)

	// Get our object's base resistances, then merge with any object-specific armors.
	armor := (o2.Resistances()).Clone()
	switch o2 := o2.(type) {
	case *ObjectCharacter:
		armor.Merge(o2.armor)
	}

	// Reduce the damages.
	for _, d := range damages {
		armor.Reduce(&d)
	}

	// Send the attacked event to the defender.
	e2 := &EventAttacked{
		Attacker: o,
		Armor:    armor,
		Damages:  damages,
	}
	o2.ResolveEvent(e2)
	if !e2.Prevented {
		fmt.Println("TODO: Apply damages: ", e2.Damages)
	}

	// Send the attack event to the attacker.
	e3 := &EventAttack{
		Target:  o2,
		Damages: damages,
	}
	o.ResolveEvent(e3)

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

	o.shouldRecalculate = true
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
	o.reach = o.CalculateReach()
	o.damages = o.CalculateDamages()
	o.armor = o.CalculateArmor()
	o.dodge = o.CalculateDodge()
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

func (o *ObjectCharacter) CalculateReach() int {
	result := 1

	// Add any bonuses from our ancestry.
	for _, a := range o.Archetype.ArchPointers {
		result += int(a.Reach)
	}
	// Add from our own archetype.
	result += int(o.AltArchetype.Reach)

	return result
}

func (o *ObjectCharacter) CalculateDamages() []Damages {
	// Get our current weapon(s).
	var weapons []*ObjectWeapon
	for _, e := range o.equipment {
		switch e := e.(type) {
		case *ObjectWeapon:
			weapons = append(weapons, e)
		}
	}
	// If we have no weapons, default to PUNCH.
	if len(weapons) == 0 {
		weapons = append(weapons, HandToHandWeapon)
	}

	var damages []Damages
	for _, w := range weapons {
		dmg, err := GetDamages(w, o)
		if err != nil {
			o.GetOwner().SendMessage(err.Error())
			continue
		}
		damages = append(damages, dmg)
	}

	return damages
}

func (o *ObjectCharacter) CalculateArmor() (armors Armors) {
	var armor *ObjectArmor
	for _, e := range o.equipment {
		switch e := e.(type) {
		case *ObjectArmor:
			armor = e
			break
		}
	}

	if armor != nil {
		armor, err := GetArmors(armor, o)
		if err != nil {
			o.GetOwner().SendMessage(err.Error())
			return
		}
		armors = armor
	}
	return
}

func (o *ObjectCharacter) CalculateDodge() (dodge float64) {
	// Get our base dodge skill
	if d, ok := o.Archetype.Skills[data.DodgeSkill]; ok {
		if c, ok := o.Archetype.Competencies[data.DodgeCompetency]; ok {
			dodge = math.Floor(d.Experience) * c.Efficiency
		}
	}
	// FIXME: Dodge should be acquired for physical, arcane, and spirit!
	// Add flat Reaction value.
	dodge += float64(o.Archetype.Attributes.Physical.GetAttribute(data.Reaction))
	dodge *= 0.0075 // Scale to 0.75% per unit.

	// Get our current worn armor value.
	var armor *ObjectArmor
	for _, e := range o.equipment {
		switch e := e.(type) {
		case *ObjectArmor:
			armor = e
			break
		}
	}

	// Reduce dodge based upon our armor if worn, or gain a 20% dodge bonus if the character is not wearing armor.
	if armor != nil && armor.Archetype.Armor > 0 {
		dodge *= 1 - armor.Archetype.Armor
	} else {
		dodge += 0.2
	}

	return
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
	d := o.GetDistance(t2.Y, t2.X, t2.Z)
	if d < 100*float64(volume) { // FIXME: Use a Sense + Focus(minor) derived value.
		o.GetOwner().SendSound(audioID, soundID, o2.GetID(), 0, 0, 0, volume)
		// TODO: EventSound?
	}
}

func (o *ObjectCharacter) Attackable() bool {
	if o.health >= 0 {
		return true
	}
	return false
}
