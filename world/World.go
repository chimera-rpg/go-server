package world

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"errors"

	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/data"
)

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
}

// Setup loads our initial starting world location and starts the
// map cleanup goroutine.
func (w *World) Setup(data *data.Manager) error {
	w.MessageChannel = make(chan MessageI)
	w.data = data
	w.players = make([]*OwnerPlayer, 0)
	w.objects = make(map[ID]ObjectI)
	w.LoadMap("Chamber of Origins")
	// FIXME: Create a temporary dummy map
	// Create a timer for doing cleanup.
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
	w.activeMapsMutex.Lock()
	for i := range w.activeMaps {
		j := i - inactivated
		if w.activeMaps[j].playerCount == 0 && w.activeMaps[j].shouldSleep == true {
			w.inactiveMaps = append(w.inactiveMaps, w.activeMaps[j])
			w.activeMaps = w.activeMaps[:j+copy(w.activeMaps[j:], w.activeMaps[j+1:])]
			inactivated++
		}
	}
	w.activeMapsMutex.Unlock()
	// Now we iterate over our inactiveMaps and remove any that have expired.
	expired := 0
	w.inactiveMapsMutex.Lock()
	for i := range w.inactiveMaps {
		j := i - expired
		if w.inactiveMaps[j].shouldExpire == true {
			w.inactiveMaps[j].Cleanup(w)
			w.inactiveMaps = w.inactiveMaps[:j+copy(w.inactiveMaps[j:], w.inactiveMaps[j+1:])]
			expired++
		}
	}
	w.inactiveMapsMutex.Unlock()
}

// Update processes updates for each player then updates each map as necessary.
func (w *World) Update(delta time.Duration) error {
	// Process world event channel.
	select {
	case msg := <-w.MessageChannel:
		switch t := msg.(type) {
		case MessageAddClient:
			if err := w.addPlayerByConnection(t.Client, t.Character); err != nil {
				log.Println("TODO: Kick player as we errored while adding them.")
			}
		case MessageRemoveClient:
			w.removePlayerByConnection(t.Client)
		default:
		}
	default:
	}
	// Process our players.
	for _, player := range w.players {
		player.Update(delta)
	}
	// Update all our active maps.
	w.activeMapsMutex.Lock()
	for _, activeMap := range w.activeMaps {
		activeMap.Update(w, delta)
	}
	w.activeMapsMutex.Unlock()
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

// addMap adds the provided Map to the active maps slice.
func (w *World) addMap(gm *Map) {
	w.activeMapsMutex.Lock()
	defer w.activeMapsMutex.Unlock()
	w.activeMaps = append(w.activeMaps, gm)
	log.WithFields(log.Fields{
		"name": gm.name,
	}).Println("Added map to active maps")
}

// activateMap activates and returns an inactive map given by its index.
func (w *World) activateMap(inactiveIndex int) *Map {
	w.inactiveMapsMutex.Lock()
	w.activeMapsMutex.Lock()
	defer w.activeMapsMutex.Unlock()
	defer w.inactiveMapsMutex.Unlock()

	if inactiveIndex > len(w.inactiveMaps) {
		log.WithFields(log.Fields{
			"index": inactiveIndex,
		}).Warnln("inactive map out of bounds")
		return nil
	}
	w.activeMaps = append(w.activeMaps, w.inactiveMaps[inactiveIndex])
	w.inactiveMaps = append(w.inactiveMaps[:inactiveIndex], w.inactiveMaps[inactiveIndex+1:]...)
	return w.activeMaps[len(w.activeMaps)-1]
}

// inactivateMap moves the given active map by index to the inactive map slice.
func (w *World) inactivateMap(activeIndex int) *Map {
	w.activeMapsMutex.Lock()
	w.inactiveMapsMutex.Lock()
	defer w.inactiveMapsMutex.Unlock()
	defer w.activeMapsMutex.Unlock()

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
	w.activeMapsMutex.Lock()
	defer w.activeMapsMutex.Unlock()
	for i := range w.activeMaps {
		if w.activeMaps[i].mapID == w.data.Strings.Acquire(name) {
			return i, true
		}
	}
	w.inactiveMapsMutex.Lock()
	defer w.inactiveMapsMutex.Unlock()
	for i := range w.inactiveMaps {
		if w.inactiveMaps[i].mapID == w.data.Strings.Acquire(name) {
			return i, false
		}
	}
	return -1, false
}

func (w *World) addPlayerByConnection(conn clientConnectionI, character *data.Character) error {
	if index := w.getExistingPlayerConnectionIndex(conn); index == -1 {
		player := NewOwnerPlayer(conn)
		conn.SetOwner(player)
		// Process and compile the character's Archetype so it inherits properly.
		w.data.ProcessArchetype(&character.Archetype)
		w.data.CompileArchetype(&character.Archetype)
		// Create character object.
		pc := NewObjectCharacterFromCharacter(character)
		pc.id = w.objectIDs.acquire()
		w.objects[pc.id] = pc
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
				gmap.AddOwner(player, 0, 1, 1)
			} else {
				return err
			}
		}
		log.WithFields(log.Fields{
			"ID": conn.GetID(),
			"PC": pc.id,
		}).Debugln("Added player to world.")
	}
	return nil
}

func (w *World) removePlayerByConnection(conn clientConnectionI) {
	if index := w.getExistingPlayerConnectionIndex(conn); index >= 0 {
		// TODO: Save ObjectCharacter to connection's associated Character data.
		// Remove owner from map -- this also deletes the character object.
		if playerMap := w.players[index].GetMap(); playerMap != nil {
			playerMap.RemoveOwner(w.players[index])
			w.DeleteObject(w.players[index].GetTarget(), true)
		}
		// Remove from our slice.
		w.players = append(w.players[:index], w.players[index+1:]...)
	}
}

func (w *World) getExistingPlayerConnectionIndex(conn clientConnectionI) int {
	for index, player := range w.players {
		if conn.GetID() == player.ClientConnection.GetID() {
			return index
		}
	}
	return -1
}

// CreateObjectFromArch will attempt to create an Object by an archetype, merging the result with the archetype's target Arch if possible.
func (w *World) CreateObjectFromArch(arch *data.Archetype) (o ObjectI, err error) {
	// Ensure archetype is compiled.
	err = w.data.CompileArchetype(arch)

	// Create our object.
	switch arch.Type {
	case cdata.ArchetypeTile:
		o = NewObjectTile(arch)
	case cdata.ArchetypeBlock:
		o = NewObjectBlock(arch)
	case cdata.ArchetypeItem:
		o = NewObjectItem(arch)
	case cdata.ArchetypePC:
		o = NewObjectCharacter(arch)
	case cdata.ArchetypeNPC:
		o = NewObjectCharacter(arch)
	case cdata.ArchetypeArmor:
		o = NewObjectArmor(arch)
	case cdata.ArchetypeShield:
		o = NewObjectShield(arch)
	case cdata.ArchetypeWeapon:
		o = NewObjectWeapon(arch)
	case cdata.ArchetypeFood:
		o = NewObjectFood(arch)
	default:
		gameobj := ObjectGeneric{
			Object: Object{
				Archetype: arch,
			},
		}
		gameobj.value, _ = arch.Value.GetInt()
		gameobj.count, _ = arch.Count.GetInt()
		gameobj.name, _ = arch.Name.GetString()

		o = &gameobj
	}
	o.SetID(w.objectIDs.acquire())
	w.objects[o.GetID()] = o

	// TODO: Create/Merge Archetype properties!
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
