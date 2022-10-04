package world

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/cosmos72/gomacro/fast"
)

// MapHandlers is our struct container all map handlers for a map.
type MapHandlers struct {
	ownerJoinFunc  MapOwnerJoinFunc
	ownerLeaveFunc MapOwnerLeaveFunc
	cleanupFunc    MapCleanupFunc
	sleepFunc      MapSleepFunc
	wakeFunc       MapWakeFunc
	updateFunc     MapUpdateFunc
	seasonFunc     MapSeasonFunc
	cycleFunc      MapCycleFunc
}

type MapOwnerJoinFunc = func(o OwnerI)
type MapOwnerLeaveFunc = func(o OwnerI)
type MapCleanupFunc = func()
type MapSleepFunc = func()
type MapWakeFunc = func()
type MapUpdateFunc = func(delta time.Duration)
type MapSeasonFunc = func(season Season)
type MapCycleFunc = func(cycle Cycle)

// setupMapHandlers sets up the gmap's handlers struct to point to the interpreter's funcs as needed.
func (gmap *Map) setupMapHandlers() {
	v := gmap.interpreter.ValueOf("OnOwnerJoin")
	if v.IsValid() {
		gmap.handlers.ownerJoinFunc = v.Interface().(MapOwnerJoinFunc)
	}
	v = gmap.interpreter.ValueOf("OnOwnerLeave")
	if v.IsValid() {
		gmap.handlers.ownerLeaveFunc = v.Interface().(MapOwnerLeaveFunc)
	}
	v = gmap.interpreter.ValueOf("OnCleanup")
	if v.IsValid() {
		gmap.handlers.cleanupFunc = v.Interface().(MapCleanupFunc)
	}
	v = gmap.interpreter.ValueOf("OnSleep")
	if v.IsValid() {
		gmap.handlers.sleepFunc = v.Interface().(MapSleepFunc)
	}
	v = gmap.interpreter.ValueOf("OnWake")
	if v.IsValid() {
		gmap.handlers.wakeFunc = v.Interface().(MapWakeFunc)
	}
	v = gmap.interpreter.ValueOf("OnUpdate")
	if v.IsValid() {
		gmap.handlers.updateFunc = v.Interface().(MapUpdateFunc)
	}
	v = gmap.interpreter.ValueOf("OnSeason")
	if v.IsValid() {
		gmap.handlers.seasonFunc = v.Interface().(MapSeasonFunc)
	}
	v = gmap.interpreter.ValueOf("OnCycle")
	if v.IsValid() {
		gmap.handlers.cycleFunc = v.Interface().(MapCycleFunc)
	}
}

// evalInterpreter attempts to evaluate the given map code.
func (gmap *Map) evalInterpreter(code string) {
	defer func() {
		if err := recover(); err != nil {
			parts := strings.Split(fmt.Sprintf("%v", err), " ")
			firstPart := parts[0]
			fileParts := strings.Split(firstPart, ":")
			lineIndex, _ := strconv.Atoi(fileParts[1])
			lineIndex--
			charIndex, _ := strconv.Atoi(fileParts[2])
			lines := strings.Split(code, "\n")

			s := strings.Join(parts[1:], " ")
			log.Errorf("Map Script: %d:%d: %s\n", lineIndex, charIndex, s)
			for i := lineIndex - 1; i <= lineIndex+1; i++ {
				if i >= 0 && i < len(lines) {
					if i == lineIndex {
						log.Errorln(fmt.Sprintf("--> %s <--", lines[i]))
					} else {
						log.Errorln(fmt.Sprintf("    %s", lines[i]))
					}
				}
			}
		}
	}()

	gmap.interpreter.Eval(code)
}

func (gmap *Map) addInterpreter(code string) {
	gmap.interpreter = fast.New()

	{
		SetupInterpreterTypes(gmap.interpreter)
		gmap.interpreter.DeclVar("gmap", nil, gmap)
		gmap.interpreter.DeclVar("world", nil, gmap.world)
		gmap.interpreter.DeclVar("data", nil, gmap.world.data)
	}

	gmap.evalInterpreter(code)

	gmap.setupMapHandlers()
}
