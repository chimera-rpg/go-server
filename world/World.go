package world

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"errors"

	"github.com/chimera-rpg/go-server/data"
	"github.com/chimera-rpg/go-server/network"
)

// FIXME: This shouldn't be here. We want to have default melee fallback, though certain genera should have alternatives that use edged or similar.
var HandToHandWeapon *ObjectEquipable

// World contains and manages all map updating, loading, and otherwise.
type World struct {
	data              *data.Manager
	activeMaps        []*Map
	activeMapsMutex   sync.Mutex
	inactiveMaps      []*Map
	inactiveMapsMutex sync.Mutex
	players           []*OwnerPlayer
	objectIDs         IDMap
	objects           map[ID]ObjectI // global objects reference.
	MessageChannel    chan MessageI
	Time              Time
}

// Setup loads our initial starting world location and starts the
// map cleanup goroutine.
func (w *World) Setup(manager *data.Manager) error {
	w.MessageChannel = make(chan MessageI)
	w.data = manager
	w.players = make([]*OwnerPlayer, 0)
	w.objects = make(map[ID]ObjectI)
	w.LoadMap("Chamber of Origins")
	// FIXME: Create a temporary dummy map
	// Create a timer for doing cleanup.
	if a, err := w.data.GetArchetypeByName("weapons/handtohand/striking"); err != nil {
		log.Errorln("couldn't load striking archetype", err)
	} else {
		if o, err := w.CreateObjectFromArch(a); err != nil {
			log.Errorln("couldn't load create striking object", err)
		} else {
			HandToHandWeapon = o.(*ObjectEquipable)
		}
	}

	cleanupTicker := time.NewTicker(time.Second * 60)
	go func() {
		for {
			<-cleanupTicker.C
			w.cleanupMaps()
		}
	}()
	return nil
}

// cleanupMaps iterates through our active and inactive maps, removing
// them as necessary.
func (w *World) cleanupMaps() {
	log.Print("Ticking map cleanup")
	// Here we iterate over our activeMaps and move any maps that should
	// enter a sleep to the inactiveMaps slices.
	inactivated := 0
	//w.activeMapsMutex.Lock()
	for i := range w.activeMaps {
		j := i - inactivated
		if w.activeMaps[j].playerCount == 0 && w.activeMaps[j].shouldSleep == true {
			w.inactiveMaps = append(w.inactiveMaps, w.activeMaps[j])
			w.activeMaps = w.activeMaps[:j+copy(w.activeMaps[j:], w.activeMaps[j+1:])]
			inactivated++
			if w.activeMaps[j].handlers.sleepFunc != nil {
				w.activeMaps[j].handlers.sleepFunc()
			}
		}
	}
	//w.activeMapsMutex.Unlock()
	// Now we iterate over our inactiveMaps and remove any that have expired.
	expired := 0
	//w.inactiveMapsMutex.Lock()
	for i := range w.inactiveMaps {
		j := i - expired
		if w.inactiveMaps[j].shouldExpire == true {
			if w.inactiveMaps[j].handlers.cleanupFunc != nil {
				w.inactiveMaps[j].handlers.cleanupFunc()
			}
			w.inactiveMaps[j].Cleanup(w)
			w.inactiveMaps = w.inactiveMaps[:j+copy(w.inactiveMaps[j:], w.inactiveMaps[j+1:])]
			expired++
		}
	}
	//w.inactiveMapsMutex.Unlock()
}

// Update processes updates for each player then updates each map as necessary.
func (w *World) Update(currentTime time.Time, delta time.Duration) error {
	updates := Updates{
		Delta: delta,
	}
	updates.Updates = append(updates.Updates, w.Time.Set(currentTime))
	// Process world event channel.
	select {
	case msg := <-w.MessageChannel:
		switch t := msg.(type) {
		case MessageAddClient:
			if err := w.addPlayerByConnection(t.Client, t.Character); err != nil {
				log.Println("TODO: Kick player as we errored while adding them.")
			}
		case MessageReplaceClient:
			w.ReplacePlayerConnection(t.Player, t.Client)
		case MessageRemoveClient:
			w.RemovePlayerByConnection(t.Client)
		default:
		}
	default:
	}
	// Process our players.
	temp := w.players[:0]
	for _, player := range w.players {
		if player.disconnected {
			player.disconnectedElapsed += delta
			// TODO: Make this influenced by map reset as well as server settings!
			if player.disconnectedElapsed > time.Duration(5)*time.Minute {
				// I guess it is okay to save the player.
				if err := w.SyncPlayerSaveInfo(player.ClientConnection); err != nil {
					log.Errorln(err)
				}
				//
				player.GetMap().RemoveOwner(player)
				w.DeleteObject(player.GetTarget(), true)
			} else {
				temp = append(temp, player)
			}
		} else {
			temp = append(temp, player)
		}
		player.Update(delta)
	}
	w.players = temp
	// Update all our active maps.
	//w.activeMapsMutex.Lock()
	for _, activeMap := range w.activeMaps {
		activeMap.Update(w, updates)
	}
	//w.activeMapsMutex.Unlock()
	return nil
}

// New returns a new World instance.
func New() *World {
	return &World{}
}

// LoadMap loads and returns a Map identified by the passed string.
func (w *World) LoadMap(name string) (*Map, error) {
	mapIndex, isActive := w.isMapLoaded(name)
	if mapIndex >= 0 {
		if !isActive {
			return w.activateMap(mapIndex), nil
		}
		return w.activeMaps[mapIndex], nil
	}
	gmap, err := NewMap(w, name)
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{
		"name": name,
	}).Println("Loaded map")

	w.addMap(gmap)

	return gmap, nil
}

// GetMap returns the a loaded map. If the map has not been loaded, this returns nil.
func (w *World) GetMap(name string) *Map {
	mapIndex, isActive := w.isMapLoaded(name)
	if mapIndex == -1 {
		return nil
	}
	if isActive {
		return w.activeMaps[mapIndex]
	}
	return w.inactiveMaps[mapIndex]
}

// RestartMap restarts the given map if it is loaded. TODO: This should send any players in this map to a Safe Place (tm) before restarting.
func (w *World) RestartMap(name string) {
	mapIndex, isActive := w.isMapLoaded(name)
	if mapIndex == -1 {
		return
	}
	if isActive {
		w.inactivateMap(mapIndex)
	}
	mapIndex, _ = w.isMapLoaded(name)
	if mapIndex == -1 {
		return
	}
	w.inactiveMaps[mapIndex].shouldExpire = true

	return
}

// addMap adds the provided Map to the active maps slice.
func (w *World) addMap(gm *Map) {
	//w.activeMapsMutex.Lock()
	//defer w.activeMapsMutex.Unlock()
	w.activeMaps = append(w.activeMaps, gm)
	log.WithFields(log.Fields{
		"name": gm.name,
	}).Println("Added map to active maps")
}

// activateMap activates and returns an inactive map given by its index.
func (w *World) activateMap(inactiveIndex int) *Map {
	//w.inactiveMapsMutex.Lock()
	//w.activeMapsMutex.Lock()
	//defer w.activeMapsMutex.Unlock()
	//defer w.inactiveMapsMutex.Unlock()

	if inactiveIndex > len(w.inactiveMaps) {
		log.WithFields(log.Fields{
			"index": inactiveIndex,
		}).Warnln("inactive map out of bounds")
		return nil
	}
	w.activeMaps = append(w.activeMaps, w.inactiveMaps[inactiveIndex])
	w.inactiveMaps = append(w.inactiveMaps[:inactiveIndex], w.inactiveMaps[inactiveIndex+1:]...)

	if w.activeMaps[len(w.activeMaps)-1].handlers.wakeFunc != nil {
		w.activeMaps[len(w.activeMaps)-1].handlers.wakeFunc()
	}

	return w.activeMaps[len(w.activeMaps)-1]
}

// inactivateMap moves the given active map by index to the inactive map slice.
func (w *World) inactivateMap(activeIndex int) *Map {
	//w.activeMapsMutex.Lock()
	//w.inactiveMapsMutex.Lock()
	//defer w.inactiveMapsMutex.Unlock()
	//defer w.activeMapsMutex.Unlock()

	if activeIndex > len(w.activeMaps) {
		log.WithFields(log.Fields{
			"index": activeIndex,
		}).Warnln("active map out of bounds")
		return nil
	}
	w.inactiveMaps = append(w.inactiveMaps, w.activeMaps[activeIndex])
	w.activeMaps = append(w.activeMaps[:activeIndex], w.activeMaps[activeIndex+1:]...)
	return w.inactiveMaps[len(w.inactiveMaps)-1]
}

// isMapLoaded returns the index and active status of a given map.
func (w *World) isMapLoaded(name string) (mapIndex int, isActive bool) {
	//w.activeMapsMutex.Lock()
	//defer w.activeMapsMutex.Unlock()
	for i := range w.activeMaps {
		if w.activeMaps[i].mapID == w.data.Strings.Acquire(name) {
			return i, true
		}
	}
	//w.inactiveMapsMutex.Lock()
	//defer w.inactiveMapsMutex.Unlock()
	for i := range w.inactiveMaps {
		if w.inactiveMaps[i].mapID == w.data.Strings.Acquire(name) {
			return i, false
		}
	}
	return -1, false
}

func (w *World) addPlayerByConnection(conn clientConnectionI, character *data.Character) error {
	if index := w.GetExistingPlayerConnectionIndex(conn); index == -1 {
		player := NewOwnerPlayer(conn)
		conn.SetOwner(player)
		// Create character object.
		pc, err := w.CreateObjectFromArch(&character.Archetype)
		if err != nil {
			return err
		}
		player.SetTarget(pc)
		// Add player to the world's record of players.
		w.players = append(w.players, player)
		// Add character object to its target map.
		if gmap, err := w.LoadMap(character.SaveInfo.Map); err == nil {
			gmap.AddOwner(player, character.SaveInfo.Y, character.SaveInfo.X, character.SaveInfo.Z)
		} else {
			log.WithFields(log.Fields{
				"name": character.SaveInfo.Map,
			}).Warnln("Could not load character's map, falling back to default")
			if gmap, err := w.LoadMap("Chamber of Origins"); err == nil {
				gmap.AddOwner(player, gmap.y, gmap.x, gmap.z)
			} else {
				return err
			}
		}
		log.WithFields(log.Fields{
			"ID": conn.GetID(),
			"PC": pc.GetID(),
		}).Debugln("Added player to world.")
		// Send player's basic inventory.
		payload := pc.(*ObjectCharacter).FeatureInventory.ToPayloadContainer()
		player.ClientConnection.Send(network.CommandObject{
			ObjectID: 0, // Object ID zero always refers to the player's basic inventory.
			Payload:  payload,
		})
		// Send inventory objects as well.
		for _, o := range pc.(*ObjectCharacter).FeatureInventory.inventory {
			player.sendObject(o)
		}
		//
		player.ClientConnection.Send(network.CommandMessage{
			Type: network.ServerMessage,
			Body: fmt.Sprintf("Welcome back, %s!", pc.Name()),
		})
		for _, p := range w.players {
			if p == player {
				continue
			}
			p.ClientConnection.Send(network.CommandMessage{
				Type: network.ServerMessage,
				Body: fmt.Sprintf("%s has materialized!", pc.Name()),
			})
		}
	}
	return nil
}

func (w *World) SyncPlayerSaveInfo(conn clientConnectionI) error {
	fmt.Println("Syncing, it would seem")
	index := w.GetExistingPlayerConnectionIndex(conn)
	if index < 0 {
		return fmt.Errorf("couldn't find player matching connection to save")
	}
	p := w.players[index]
	u := conn.GetUser()
	if u == nil {
		return fmt.Errorf("couldn't find connection's user to save")
	}
	o := p.GetTarget()
	if o == nil {
		return fmt.Errorf("user %s has no target to save", u.Username)
	}
	t := o.GetTile()
	if t == nil {
		return fmt.Errorf("player object %s has no owning tile", o.Name())
	}
	m := t.GetMap()
	if m == nil {
		return fmt.Errorf("player object %s's tile has no map", o.Name())
	}
	if _, ok := u.Characters[o.Name()]; !ok {
		return fmt.Errorf("user '%s' has no matching character '%s'", u.Username, o.Name())
	}
	s := u.Characters[o.Name()].SaveInfo
	s.Map = m.dataName
	s.X = t.X
	s.Y = t.Y
	s.Z = t.Z
	s.Time = time.Now()
	if m.haven || t.haven {
		s.HavenMap = m.dataName
		s.HavenX = t.X
		s.HavenY = t.Y
		s.HavenZ = t.Z
	}
	u.Characters[o.Name()].SaveInfo = s
	u.Characters[o.Name()].Archetype = o.GetSaveableArchetype()
	fmt.Println("Set SaveInfo")
	return nil
}
func (w *World) SavePlayerByUsername(username string) error {
	p := w.GetPlayerByUsername(username)
	if p == nil {
		return fmt.Errorf("couldn't find username to save: %s", username)
	}
	t := p.GetTarget()
	if t == nil {
		return fmt.Errorf("username %s has no target to save", username)
	}

	return nil
}

// ReplacePlayerConnection replaces the connection for the given player.
func (w *World) ReplacePlayerConnection(player *OwnerPlayer, conn clientConnectionI) {
	// Reset view and known ids
	player.CreateView()
	player.knownIDs = make(map[uint32]struct{})
	// Update network stuff
	player.disconnected = false
	player.disconnectedElapsed = 0
	player.ClientConnection = conn

	// Refresh the client's target object.
	player.ClientConnection.Send(network.CommandObject{
		ObjectID: player.GetTarget().GetID(),
		Payload: network.CommandObjectPayloadViewTarget{
			Height: uint8(player.viewHeight),
			Width:  uint8(player.viewWidth),
			Depth:  uint8(player.viewDepth),
		},
	})

	// Send character's base inventory.
	player.ClientConnection.Send(network.CommandObject{
		ObjectID: 0, // Object ID zero always refers to the player's basic inventory.
		Payload:  player.GetTarget().(*ObjectCharacter).FeatureInventory.ToPayloadContainer(),
	})
	// Send inventory objects as well.
	for _, o := range player.GetTarget().(*ObjectCharacter).FeatureInventory.inventory {
		player.sendObject(o)
	}

	// SetMap causes the initial map command to be sent, as well as resetting known IDs and view.
	player.SetMap(player.currentMap)

}

// RemovePlayerByConnection does as it implies.
func (w *World) RemovePlayerByConnection(conn clientConnectionI) {
	if index := w.GetExistingPlayerConnectionIndex(conn); index >= 0 {
		// Note, we do not remove the player if the player's target is not in a haven.
		player := w.players[index]
		if w.IsPlayerInHaven(player) {
			w.RemovePlayerByIndex(index)
		} else {
			w.players[index].disconnected = true
			w.players[index].disconnectedElapsed = 0
			// Replace connection with a dummy one if not in haven.
			w.players[index].ClientConnection = &dummyConnection{
				user:  player.ClientConnection.GetUser(),
				id:    player.ClientConnection.GetID(),
				owner: player.ClientConnection.GetOwner(),
			}
		}
	}
}

func (w *World) RemovePlayerByIndex(index int) {
	player := w.players[index]
	player.GetMap().RemoveOwner(player)
	w.DeleteObject(player.GetTarget(), true)
	w.players = append(w.players[:index], w.players[index+1:]...)
}

func (w *World) IsPlayerInHaven(player *OwnerPlayer) bool {
	if playerMap := player.GetMap(); playerMap != nil {
		if playerMap.haven || player.GetTarget().GetTile().haven {
			return true
		}
	}
	return false
}

func (w *World) GetExistingPlayerConnectionIndex(conn clientConnectionI) int {
	for index, player := range w.players {
		if conn.GetID() == player.ClientConnection.GetID() {
			return index
		}
	}
	return -1
}

// CreateObject looks up an archetype matching a string and then calls CreateObjectFromArch.
func (w *World) CreateObject(s string) (o ObjectI, err error) {
	if a, err := w.data.GetArchetypeByName(s); err != nil {
		return nil, err
	} else {
		return w.CreateObjectFromArch(a)
	}
}

// CreateObjectFromArch will attempt to create an Object by an archetype, merging the result with the archetype's target Arch if possible.
func (w *World) CreateObjectFromArch(arch *data.Archetype) (o ObjectI, err error) {
	// Ensure archetype is processed.
	err = w.data.ProcessArchetype(arch)
	// Ensure archetype is compiled.
	err = w.data.CompileArchetype(arch)

	// Create our object.
	switch arch.Type {
	case data.ArchetypeTile:
		o = NewObjectTile(arch)
	case data.ArchetypeBlock:
		o = NewObjectBlock(arch)
	case data.ArchetypeItem:
		o = NewObjectItem(arch)
	case data.ArchetypeVariety:
		o = NewObjectCharacter(arch)
	case data.ArchetypePC:
		o = NewObjectCharacter(arch)
	case data.ArchetypeNPC:
		o = NewObjectCharacter(arch)
	case data.ArchetypeEquipable:
		o = NewObjectEquipable(arch)
	case data.ArchetypeFood:
		o = NewObjectFood(arch)
	case data.ArchetypeAudio:
		o = NewObjectAudio(arch)
	case data.ArchetypeSpecial:
		o = NewObjectSpecial(arch)
	case data.ArchetypeFlora:
		o = NewObjectFlora(arch)
	case data.ArchetypeExit:
		o = NewObjectExit(arch)
	default:
		gameobj := ObjectGeneric{
			Object: NewObject(arch),
		}
		// TODO: Create/use a simple scripting language for rolling dynamic values.
		if arch.Value != nil {
			if i, err := strconv.Atoi(*arch.Value); err == nil {
				gameobj.value = i
			}
		}
		if arch.Count != nil {
			if i, err := strconv.Atoi(*arch.Count); err == nil {
				gameobj.count = i
			}
		}
		if arch.Name != "" {
			gameobj.name = arch.Name
		}

		o = &gameobj
	}
	o.SetID(w.objectIDs.acquire())
	w.objects[o.GetID()] = o

	// FIXME: Should all objects implement inventory?
	for _, invArch := range arch.Inventory {
		invObj, err := w.CreateObjectFromArch(&invArch)
		if err != nil {
			return o, err
		}
		switch o := o.(type) {
		case *ObjectCharacter:
			o.AddInventoryObject(invObj)
			invObj.SetContainer(o)
		}
	}
	for _, eqArch := range arch.Equipment {
		eqObj, err := w.CreateObjectFromArch(&eqArch)
		if err != nil {
			return o, err
		}
		switch o := o.(type) {
		case *ObjectCharacter:
			o.AddEquipmentObject(eqObj)
		}
	}

	return
}

// DeleteObject deletes a given object. If shouldFree is true, the associated object ID is freed.
func (w *World) DeleteObject(o ObjectI, shouldFree bool) (err error) {
	if o == nil {
		return errors.New("attempted to delete a nil object")
	}
	if tile := o.GetTile(); tile != nil {
		if m := tile.GetMap(); m != nil {
			err = m.RemoveObject(o)
		}
	}
	if shouldFree {
		w.objectIDs.free(o.GetID())
		delete(w.objects, o.GetID())
	}

	switch o := o.(type) {
	case *ObjectCharacter:
		for _, o2 := range o.inventory {
			w.DeleteObject(o2, true) // Is it proper to free?
		}
		for _, o2 := range o.equipment {
			w.DeleteObject(o2, true)
		}
	}

	return
}

// GetObject gets an ObjectI if it exists.
func (w *World) GetObject(oID ID) ObjectI {
	return w.objects[oID]
}

// GetPlayers returns a slice of all active players.
func (w *World) GetPlayers() []*OwnerPlayer {
	return w.players
}

// GetPlayerByUsername returns a player by their owning user name.
func (w *World) GetPlayerByUsername(name string) *OwnerPlayer {
	for _, p := range w.players {
		if p.ClientConnection.GetUser().Username == name {
			return p
		}
	}
	return nil
}

// GetPlayerByObjectID returns a player by their object id.
func (w *World) GetPlayerByObjectID(oID ID) *OwnerPlayer {
	for _, p := range w.players {
		if p.target.id == oID {
			return p
		}
	}
	return nil
}
