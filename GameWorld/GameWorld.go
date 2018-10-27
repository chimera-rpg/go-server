package GameWorld

import (
  "log"
  "time"
  "sync"
  "server/GameData"
)

type GameWorld struct {
  gameData          *GameData.Manager
  activeMaps        []*GameMap
  activeMapsMutex   sync.Mutex
  inactiveMaps      []*GameMap
  inactiveMapsMutex sync.Mutex
  players           []*OwnerPlayer
}

func (world *GameWorld) Setup(data *GameData.Manager) error {
  world.gameData = data
  world.LoadMap("Chamber of Origins")
  // FIXME: Create a temporary dummy map
  // Create a timer for doing cleanup.
  cleanup_ticker := time.NewTicker(time.Second * 60)
  go func() {
    for {
      <- cleanup_ticker.C
      world.cleanupMaps()
    }
  }()
  return nil
}

func (world *GameWorld) cleanupMaps() {
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

func (world *GameWorld) Update(delta int64) error {
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

func New() *GameWorld {
  return &GameWorld{
  }
}

func (world *GameWorld) LoadMap(name string) (*GameMap, error) {
  log.Printf("Attempting to load map '%s'\n", name)
  mapIndex, isActive := world.isMapLoaded(name)
  if mapIndex >= 0 {
    if !isActive {
      return world.activateMap(mapIndex), nil
    } else {
      return world.inactivateMap(mapIndex), nil
    }
  }
  gmap, err := NewGameMap(world.gameData, "Chamber of Origins")
  if err != nil {
    return nil, err
  }
  world.addMap(gmap)

  return gmap, nil
}

func (world *GameWorld) addMap(gm *GameMap) {
  log.Printf("Added map '%s' to active maps\n", gm.name)
  world.activeMapsMutex.Lock()
  defer world.activeMapsMutex.Unlock()
  world.activeMaps = append(world.activeMaps, gm)
}

func (world *GameWorld) activateMap(inactiveIndex int) *GameMap {
  world.inactiveMapsMutex.Lock()
  world.activeMapsMutex.Lock()
  defer world.activeMapsMutex.Unlock()
  defer world.inactiveMapsMutex.Unlock()

  if inactiveIndex > len(world.inactiveMaps) {
    log.Printf("inactive map '%s' out of bounds.\n", inactiveIndex)
    return nil
  }
  world.activeMaps = append(world.activeMaps, world.inactiveMaps[inactiveIndex])
  world.inactiveMaps = append(world.inactiveMaps[:inactiveIndex], world.inactiveMaps[inactiveIndex+1:]...)
  return world.activeMaps[len(world.activeMaps)-1]
}

func (world *GameWorld) inactivateMap(activeIndex int) *GameMap {
  world.activeMapsMutex.Lock()
  world.inactiveMapsMutex.Lock()
  defer world.inactiveMapsMutex.Unlock()
  defer world.activeMapsMutex.Unlock()

  if activeIndex > len(world.activeMaps) {
    log.Printf("active map '%s' out of bounds.\n", activeIndex)
    return nil
  }
  world.inactiveMaps = append(world.inactiveMaps, world.activeMaps[activeIndex])
  world.activeMaps = append(world.activeMaps[:activeIndex], world.activeMaps[activeIndex+1:]...)
  return world.inactiveMaps[len(world.inactiveMaps)-1]
}

func (world *GameWorld) isMapLoaded(name string) (mapIndex int, isActive bool) {
  world.activeMapsMutex.Lock()
  defer world.activeMapsMutex.Unlock()
  for i := range world.activeMaps {
    if (world.activeMaps[i].dataName == name) {
      return i, true
    }
  }
  world.inactiveMapsMutex.Lock()
  defer world.inactiveMapsMutex.Unlock()
  for i := range world.inactiveMaps {
    if (world.inactiveMaps[i].dataName == name) {
      return i, false
    }
  }
  return -1, false
}
