package world

import (
	"log"
	"sync"
	"time"

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
	MessageChannel    chan MessageI
}

// Setup loads our initial starting world location and starts the
// map cleanup goroutine.
func (world *World) Setup(data *data.Manager) error {
	world.MessageChannel = make(chan MessageI)
	world.data = data
	world.players = make([]*OwnerPlayer, 0)
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
			world.inactiveMaps = world.inactiveMaps[:j+copy(world.inactiveMaps[j:], world.inactiveMaps[j+1:])]
			expired++
		}
	}
	world.inactiveMapsMutex.Unlock()
}

// Update processes updates for each player then updates each map as necessary.
func (world *World) Update(delta int64) error {
	// Process world event channel.
	select {
	case msg := <-world.MessageChannel:
		switch t := msg.(type) {
		case MessageAddClient:
			world.addPlayerByConnection(t.Client, t.Character)
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
	log.Printf("Attempting to load map '%s'\n", name)
	mapIndex, isActive := world.isMapLoaded(name)
	if mapIndex >= 0 {
		if !isActive {
			return world.activateMap(mapIndex), nil
		}
		return world.inactivateMap(mapIndex), nil
	}
	gmap, err := NewMap(world.data, name)
	if err != nil {
		return nil, err
	}
	world.addMap(gmap)

	return gmap, nil
}

// addMap adds the provided Map to the active maps slice.
func (world *World) addMap(gm *Map) {
	log.Printf("Added map '%s' to active maps\n", gm.name)
	world.activeMapsMutex.Lock()
	defer world.activeMapsMutex.Unlock()
	world.activeMaps = append(world.activeMaps, gm)
}

// activateMap activates and returns an inactive map given by its index.
func (world *World) activateMap(inactiveIndex int) *Map {
	world.inactiveMapsMutex.Lock()
	world.activeMapsMutex.Lock()
	defer world.activeMapsMutex.Unlock()
	defer world.inactiveMapsMutex.Unlock()

	if inactiveIndex > len(world.inactiveMaps) {
		log.Printf("inactive map '%d' out of bounds.\n", inactiveIndex)
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
		log.Printf("active map '%d' out of bounds.\n", activeIndex)
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

func (world *World) addPlayerByConnection(conn clientConnectionI, character *data.Character) {
	if index := world.getExistingPlayerConnectionIndex(conn); index == -1 {
		player := NewOwnerPlayer(conn)
		// Create character object.
		pc := NewObjectPCFromCharacter(character)
		player.SetTarget(pc)
		// Add player to the world's record of players.
		world.players = append(world.players, player)
		// Add character object to its target map. TODO: Read target map from character and use fallback if map does not exist.
		if gmap, err := world.LoadMap("Chamber of Origins"); err == nil {
			gmap.PlaceObject(pc, 0, 1, 1)
		} else {
			log.Println("Could not load character's map")
		}
		log.Println("Added player and PC to world.")
	}
}

func (world *World) removePlayerByConnection(conn clientConnectionI) {
	if index := world.getExistingPlayerConnectionIndex(conn); index >= 0 {
		// Remove character object from its owning tile.
		if tile := world.players[index].target.GetTile(); tile != nil {
			tile.removeObject(world.players[index].target)
		}
		// TODO: Save ObjectPC to connection's associated Character data.
		world.players = append(world.players[:index], world.players[index+1:]...)
		log.Println("Removed player and PC from world.")
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
