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
}

// Setup loads our initial starting world location and starts the
// map cleanup goroutine.
func (world *World) Setup(data *data.Manager) error {
	world.data = data
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
	gmap, err := NewMap(world.data, "Chamber of Origins")
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
