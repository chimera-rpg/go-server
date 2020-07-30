package world

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"time"

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
func (world *World) Setup(data *data.Manager) error {
	world.MessageChannel = make(chan MessageI)
	world.data = data
	world.players = make([]*OwnerPlayer, 0)
	world.objects = make(map[ID]ObjectI)
	world.LoadMap("Chamber of Origins")
	// FIXME: Create a temporary dummy map
	// Create a timer for doing cleanup.
	cleanupTicker := time.NewTicker(time.Second * 60)
	go func() {
		for {
			<-cleanupTicker.C
			world.cleanupMaps()
		}
	}()
	return nil
}

// cleanupMaps iterates through our active and inactive maps, removing
// them as necessary.
func (world *World) cleanupMaps() {
	log.Print("Ticking map cleanup")
	// Here we iterate over our activeMaps and move any maps that should
	// enter a sleep to the inactiveMaps slices.
	inactivated := 0
	world.activeMapsMutex.Lock()
	for i := range world.activeMaps {
		j := i - inactivated
		if world.activeMaps[j].playerCount == 0 && world.activeMaps[j].shouldSleep == true {
			world.inactiveMaps = append(world.inactiveMaps, world.activeMaps[j])
			world.activeMaps = world.activeMaps[:j+copy(world.activeMaps[j:], world.activeMaps[j+1:])]
			inactivated++
		}
	}
	world.activeMapsMutex.Unlock()
	// Now we iterate over our inactiveMaps and remove any that have expired.
	expired := 0
	world.inactiveMapsMutex.Lock()
	for i := range world.inactiveMaps {
		j := i - expired
		if world.inactiveMaps[j].shouldExpire == true {
			world.inactiveMaps[j].Cleanup(world)
			world.inactiveMaps = world.inactiveMaps[:j+copy(world.inactiveMaps[j:], world.inactiveMaps[j+1:])]
			expired++
		}
	}
	world.inactiveMapsMutex.Unlock()
}

// Update processes updates for each player then updates each map as necessary.
func (world *World) Update(delta time.Duration) error {
	// Process world event channel.
	select {
	case msg := <-world.MessageChannel:
		switch t := msg.(type) {
		case MessageAddClient:
			if err := world.addPlayerByConnection(t.Client, t.Character); err != nil {
				log.Println("TODO: Kick player as we errored while adding them.")
			}
		case MessageRemoveClient:
			world.removePlayerByConnection(t.Client)
		default:
		}
	default:
	}
	// Process our players.
	for _, player := range world.players {
		player.Update(delta)
	}
	// Update all our active maps.
	world.activeMapsMutex.Lock()
	for _, activeMap := range world.activeMaps {
		activeMap.Update(world, delta)
	}
	world.activeMapsMutex.Unlock()
	return nil
}

// New returns a new World instance.
func New() *World {
	return &World{}
}

// LoadMap loads and returns a Map identified by the passed string.
func (world *World) LoadMap(name string) (*Map, error) {
	mapIndex, isActive := world.isMapLoaded(name)
	if mapIndex >= 0 {
		if !isActive {
			return world.activateMap(mapIndex), nil
		}
		return world.activeMaps[mapIndex], nil
	}
	gmap, err := NewMap(world, name)
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{
		"name": name,
	}).Println("Loaded map")

	world.addMap(gmap)

	return gmap, nil
}

// GetMap returns the a loaded map. If the map has not been loaded, this returns nil.
func (world *World) GetMap(name string) *Map {
	mapIndex, isActive := world.isMapLoaded(name)
	if mapIndex == -1 {
		return nil
	}
	if isActive {
		return world.activeMaps[mapIndex]
	} else {
		return world.inactiveMaps[mapIndex]
	}
	return nil
}

// addMap adds the provided Map to the active maps slice.
func (world *World) addMap(gm *Map) {
	world.activeMapsMutex.Lock()
	defer world.activeMapsMutex.Unlock()
	world.activeMaps = append(world.activeMaps, gm)
	log.WithFields(log.Fields{
		"name": gm.name,
	}).Println("Added map to active maps")
}

// activateMap activates and returns an inactive map given by its index.
func (world *World) activateMap(inactiveIndex int) *Map {
	world.inactiveMapsMutex.Lock()
	world.activeMapsMutex.Lock()
	defer world.activeMapsMutex.Unlock()
	defer world.inactiveMapsMutex.Unlock()

	if inactiveIndex > len(world.inactiveMaps) {
		log.WithFields(log.Fields{
			"index": inactiveIndex,
		}).Warnln("inactive map out of bounds")
		return nil
	}
	world.activeMaps = append(world.activeMaps, world.inactiveMaps[inactiveIndex])
	world.inactiveMaps = append(world.inactiveMaps[:inactiveIndex], world.inactiveMaps[inactiveIndex+1:]...)
	return world.activeMaps[len(world.activeMaps)-1]
}

// inactivateMap moves the given active map by index to the inactive map slice.
func (world *World) inactivateMap(activeIndex int) *Map {
	world.activeMapsMutex.Lock()
	world.inactiveMapsMutex.Lock()
	defer world.inactiveMapsMutex.Unlock()
	defer world.activeMapsMutex.Unlock()

	if activeIndex > len(world.activeMaps) {
		log.WithFields(log.Fields{
			"index": activeIndex,
		}).Warnln("active map out of bounds")
		return nil
	}
	world.inactiveMaps = append(world.inactiveMaps, world.activeMaps[activeIndex])
	world.activeMaps = append(world.activeMaps[:activeIndex], world.activeMaps[activeIndex+1:]...)
	return world.inactiveMaps[len(world.inactiveMaps)-1]
}

// isMapLoaded returns the index and active status of a given map.
func (world *World) isMapLoaded(name string) (mapIndex int, isActive bool) {
	world.activeMapsMutex.Lock()
	defer world.activeMapsMutex.Unlock()
	for i := range world.activeMaps {
		if world.activeMaps[i].mapID == world.data.Strings.Acquire(name) {
			return i, true
		}
	}
	world.inactiveMapsMutex.Lock()
	defer world.inactiveMapsMutex.Unlock()
	for i := range world.inactiveMaps {
		if world.inactiveMaps[i].mapID == world.data.Strings.Acquire(name) {
			return i, false
		}
	}
	return -1, false
}

func (world *World) addPlayerByConnection(conn clientConnectionI, character *data.Character) error {
	if index := world.getExistingPlayerConnectionIndex(conn); index == -1 {
		player := NewOwnerPlayer(conn)
		conn.SetOwner(player)
		// Process and compile the character's Archetype so it inherits properly.
		world.data.ProcessArchetype(&character.Archetype)
		world.data.CompileArchetype(&character.Archetype)
		// Create character object.
		pc := NewObjectCharacterFromCharacter(character)
		pc.id = world.objectIDs.acquire()
		world.objects[pc.id] = pc
		player.SetTarget(pc)
		// Add player to the world's record of players.
		world.players = append(world.players, player)
		// Add character object to its target map.
		if gmap, err := world.LoadMap(character.SaveInfo.Map); err == nil {
			gmap.AddOwner(player, character.SaveInfo.Y, character.SaveInfo.X, character.SaveInfo.Z)
		} else {
			log.WithFields(log.Fields{
				"name": character.SaveInfo.Map,
			}).Warnln("Could not load character's map, falling back to default")
			if gmap, err := world.LoadMap("Chamber of Origins"); err == nil {
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

func (world *World) removePlayerByConnection(conn clientConnectionI) {
	if index := world.getExistingPlayerConnectionIndex(conn); index >= 0 {
		// TODO: Save ObjectCharacter to connection's associated Character data.
		// Remove owner from map -- this also deletes the character object.
		if playerMap := world.players[index].GetMap(); playerMap != nil {
			playerMap.RemoveOwner(world.players[index])
			world.DeleteObject(world.players[index].GetTarget(), true)
		}
		// Remove from our slice.
		world.players = append(world.players[:index], world.players[index+1:]...)
	}
}

func (world *World) getExistingPlayerConnectionIndex(conn clientConnectionI) int {
	for index, player := range world.players {
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
	case cdata.ArchetypeCharacter:
		o = NewObjectCharacter(arch)
	case cdata.ArchetypeArmor:
		o = NewObjectArmor(arch)
	case cdata.ArchetypeShield:
		o = NewObjectShield(arch)
	case cdata.ArchetypeWeapon:
		o = NewObjectWeapon(arch)
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
		return errors.New("Attempted to delete a nil object!")
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

// GetPlayerByName returns a player by their owning user name.
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
