package data

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"

	cdata "github.com/chimera-rpg/go-common/data"
	"github.com/chimera-rpg/go-server/config"
)

// Manager is the controlling type for accessing archetypes, maps, player
// files and any other data.
type Manager struct {
	dataPath       string
	archetypesPath string
	audioPath      string
	varPath        string
	etcPath        string
	usersPath      string
	usersMutex     sync.Mutex
	//musicPath string
	//soundPath string
	mapsPath         string
	AnimationsConfig cdata.AnimationsConfig
	archetypes       map[StringID]*Archetype // Full Map of archetypes.
	//animations map[string]*Animation // Full Map of animations.
	animations map[StringID]*Animation // ID to Animation map
	audio      map[StringID]*Audio
	// Hmm... almost map[uint32]*Archetype... with CRC id
	Strings      Strings
	imageFileMap FileMap
	soundFileMap FileMap
	// images map[string][]bytes
	generaArchetypes  []*Archetype     // Slice of genera archetypes.
	speciesArchetypes []*Archetype     // Slice of species archetypes.
	pcArchetypes      []*Archetype     // Player Character archetypes, used for creating new characters.
	factionArchetypes []*Archetype     // Faction archetypes, used for looking up default attitudes.
	maps              map[string]*Map  // Full map of Maps.
	loadedUsers       map[string]*User // Map of loaded Players
	cryptParams       cryptParams      // Cryptography parameters
	// FIXME: Made these exported because I'm lazy.
	TypeHints map[StringID]string
	Slots     map[StringID]string
}

type objectTemplate struct {
	variables map[string]objectVariable
}

type objectVariable struct {
	key   string
	value string
}

// parse, process, compile
func (m *Manager) parseArchetypeFile(filepath string) error {
	defer func() {
		if err := recover(); err != nil {
			switch err := err.(type) {
			case ScriptError:
				log.Errorf("Script: %s %d:%d: %s\n", filepath, err.lineIndex, err.charIndex, err.s)
				for _, s := range err.lines {
					log.Errorln(s)
				}
			}
		}
	}()
	r, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	archetypesMap := make(map[string]*Archetype)

	if err = yaml.Unmarshal(r, &archetypesMap); err != nil {
		return err
	}
	for k, archetype := range archetypesMap {
		archID := m.Strings.Acquire(k)
		m.archetypes[archID] = archetype
		m.archetypes[archID].SelfID = archID
	}
	return nil
}

// ProcessArchetype converts certain fields of an Archetype into optimized versions. This converts Anim, Face, Arch, and Archs to their ID representation. This also processes any Inventory archetypes.
func (m *Manager) ProcessArchetype(archetype *Archetype) error {
	if archetype.isProcessing {
		return nil
	}
	archetype.isProcessing = true
	if archetype.Anim != "" {
		archetype.AnimID = m.Strings.Acquire(archetype.Anim)
		archetype.Anim = ""
	}
	if archetype.Face != "" {
		archetype.FaceID = m.Strings.Acquire(archetype.Face)
		archetype.Face = ""
	}
	// Add Arch to Archs if defined.
	if archetype.Arch != "" {
		archetype.Archs = append(archetype.Archs, archetype.Arch)
		archetype.Arch = ""
	}

	// Process audio
	if archetype.Audio != "" {
		archetype.AudioID = m.Strings.Acquire(archetype.Audio)
		archetype.Audio = ""
	}
	if archetype.SoundSet != "" {
		archetype.SoundSetID = m.Strings.Acquire(archetype.SoundSet)
		archetype.SoundSet = ""
	}

	// Convert Archs into ArchIDs
	for _, archname := range archetype.Archs {
		isAdd := false
		if archname[0] == '+' {
			archname = archname[1:]
			isAdd = true
		}
		targetID := m.Strings.Acquire(archname)
		ancestorArchetype, err := m.GetArchetype(targetID)
		if err != nil {
			return fmt.Errorf("\"%s\" does not exist", archname)
		}
		mergeArch := MergeArch{
			ID:   targetID,
			Type: ArchMerge,
		}
		if isAdd {
			mergeArch.Type = ArchAdd
		}
		archetype.ArchIDs = append(archetype.ArchIDs, mergeArch)
		archetype.ArchPointers = append(archetype.ArchPointers, ancestorArchetype)
	}
	//archetype.Archs = nil // Might as well keep the Archs references, I suppose

	// Process Inventory.
	for i := range archetype.Inventory {
		if err := m.ProcessArchetype(&archetype.Inventory[i]); err != nil {
			return err
		}
	}

	// Process type hint
	for _, v := range archetype.TypeHints {
		archetype.TypeHintIDs = append(archetype.TypeHintIDs, m.Strings.Acquire(v))
		m.TypeHints[m.Strings.Acquire(v)] = v
	}

	// Process Slots. FIXME: This parsing of all strings for sending slot information feels awful.
	for k, v := range archetype.Slots.Has {
		slot := m.Strings.Acquire(k)
		archetype.Slots.HasIDs[slot] = v
		m.Slots[slot] = k
	}
	for k, v := range archetype.Slots.Uses {
		slot := m.Strings.Acquire(k)
		archetype.Slots.UsesIDs[slot] = v
		m.Slots[slot] = k
	}
	for k, v := range archetype.Slots.Gives {
		slot := m.Strings.Acquire(k)
		archetype.Slots.GivesIDs[slot] = v
		m.Slots[slot] = k
	}
	for k, v := range archetype.Slots.Needs.Min {
		slot := m.Strings.Acquire(k)
		archetype.Slots.Needs.MinIDs[slot] = v
		m.Slots[slot] = k
	}
	for k, v := range archetype.Slots.Needs.Max {
		slot := m.Strings.Acquire(k)
		archetype.Slots.Needs.MaxIDs[slot] = v
		m.Slots[slot] = k
	}

	// Process Events' archetypes.
	processEventResponses := func(e *EventResponses) error {
		if e == nil {
			return nil
		}
		if e.Spawn != nil {
			for _, a := range e.Spawn.Items {
				if err := m.ProcessArchetype(a.Archetype); err != nil {
					return err
				}
			}
		}
		if e.Replace != nil {
			for _, a := range *e.Replace {
				if err := m.ProcessArchetype(a.Archetype); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if archetype.Events != nil {
		if err := processEventResponses(archetype.Events.Birth); err != nil {
			return err
		}
		if err := processEventResponses(archetype.Events.Death); err != nil {
			return err
		}
		if err := processEventResponses(archetype.Events.Advance); err != nil {
			return err
		}
		if err := processEventResponses(archetype.Events.Hit); err != nil {
			return err
		}
	}

	return nil
}

// CompileArchetype compiles a given archetype if it has not been compiled yet. This handles dependency resolution and will throw if an archetype has circular dependencies.
func (m *Manager) CompileArchetype(archetype *Archetype) error {
	// Bail early if already compiled or is compiling.
	if archetype.isCompiled || archetype.isCompiling {
		return nil
	}
	archetype.isCompiling = true

	// Ensure there are no circular deps.
	err := m.resolveArchetype(archetype)
	if err != nil {
		return err
	}

	// Ensure deps are all compiled and inherit linearly.
	for _, dep := range archetype.ArchIDs {
		depArch, err := m.GetArchetype(dep.ID)
		if err != nil {
			return err
		}
		if err := m.CompileArchetype(depArch); err != nil {
			return err
		}
		if dep.Type == ArchMerge {
			if err := archetype.Merge(depArch); err != nil {
				return err
			}
		} else if dep.Type == ArchAdd {
			if err := archetype.Add(depArch); err != nil {
				return err
			}
		}
	}

	// Ensure inventory is compiled.
	for i := range archetype.Inventory {
		if err := m.CompileArchetype(&archetype.Inventory[i]); err != nil {
			return err
		}
	}

	// Ensure Events' archetypes are compiled.
	compileEventResponses := func(e *EventResponses) {
		if e == nil {
			return
		}
		if e.Spawn != nil {
			for _, a := range e.Spawn.Items {
				m.CompileArchetype(a.Archetype)
			}
		}
		if e.Replace != nil {
			for _, a := range *e.Replace {
				m.CompileArchetype(a.Archetype)
			}
		}
	}
	if archetype.Events != nil {
		compileEventResponses(archetype.Events.Birth)
		compileEventResponses(archetype.Events.Death)
		compileEventResponses(archetype.Events.Advance)
		compileEventResponses(archetype.Events.Hit)
	}

	archetype.isCompiled = true

	return nil
}

func (m *Manager) resolveArchetype(archetype *Archetype) error {
	resolved := make(map[StringID]struct{})
	unresolved := make(map[StringID]struct{})

	if err := m.dependencyResolveArchetype(archetype, resolved, unresolved); err != nil {
		return err
	}
	return nil
}

func (m *Manager) dependencyResolveArchetype(archetype *Archetype, resolved, unresolved map[StringID]struct{}) error {
	unresolved[archetype.SelfID] = struct{}{}
	for _, dep := range archetype.ArchIDs {
		depArch, err := m.GetArchetype(dep.ID)
		if err != nil {
			return err
		}
		if _, ok := resolved[dep.ID]; !ok {
			if _, ok := unresolved[dep.ID]; ok {
				return fmt.Errorf("circular dependency between %s and %s", m.Strings.Lookup(archetype.SelfID), m.Strings.Lookup(dep.ID))
			}
			if err := m.dependencyResolveArchetype(depArch, resolved, unresolved); err != nil {
				return err
			}
		}
	}
	resolved[archetype.SelfID] = struct{}{}
	delete(unresolved, archetype.SelfID)

	return nil
}

func (m *Manager) parseArchetypeFiles() error {
	l := log.WithFields(log.Fields{
		"path": m.archetypesPath,
	})
	l.Print("Archetypes: Loading...")
	err := filepath.Walk(m.archetypesPath, func(filepath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if strings.HasSuffix(filepath, ".arch.yaml") {
				err = m.parseArchetypeFile(filepath)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	// Post-process our archetypes so they properly set up inheritance relationships.
	for _, archetype := range m.archetypes {
		if err := m.ProcessArchetype(archetype); err != nil {
			return err
		}
	}
	for _, archetype := range m.archetypes {
		if err := m.resolveArchetype(archetype); err != nil {
			return err
		}
	}
	for _, archetype := range m.archetypes {
		if err := m.CompileArchetype(archetype); err != nil {
			return err
		}
	}

	m.buildGeneraArchetypes()
	m.buildSpeciesArchetypes()
	m.buildFactionArchetypes()
	m.buildPCArchetypes()

	l.WithFields(log.Fields{
		"Total":      len(m.archetypes),
		"Genera":     len(m.generaArchetypes),
		"Species":    len(m.speciesArchetypes),
		"Factions":   len(m.factionArchetypes),
		"Characters": len(m.pcArchetypes),
	}).Println("Archetypes: Done!")

	return nil
}

func (m *Manager) buildPCArchetypes() int {
	oldCount := len(m.pcArchetypes)
	for _, v := range m.archetypes {
		if v.Type == cdata.ArchetypePC {
			m.pcArchetypes = append(m.pcArchetypes, v)
		}
	}
	return len(m.pcArchetypes) - oldCount
}

func (m *Manager) buildGeneraArchetypes() int {
	oldCount := len(m.generaArchetypes)
	for _, v := range m.archetypes {
		if v.Type == cdata.ArchetypeGenus {
			m.generaArchetypes = append(m.generaArchetypes, v)
		}
	}
	return len(m.generaArchetypes) - oldCount
}

func (m *Manager) buildSpeciesArchetypes() int {
	oldCount := len(m.speciesArchetypes)
	for _, v := range m.archetypes {
		if v.Type == cdata.ArchetypeSpecies {
			m.speciesArchetypes = append(m.speciesArchetypes, v)
		}
	}
	return len(m.speciesArchetypes) - oldCount
}

func (m *Manager) buildFactionArchetypes() int {
	oldCount := len(m.factionArchetypes)
	for _, v := range m.archetypes {
		if v.Type == cdata.ArchetypeFaction {
			m.factionArchetypes = append(m.factionArchetypes, v)
		}
	}
	return len(m.factionArchetypes) - oldCount
}

// GetArchetype gets the given archetype by id if it exists.
func (m *Manager) GetArchetype(archID StringID) (archetype *Archetype, err error) {
	if _, ok := m.archetypes[archID]; ok {
		return m.archetypes[archID], nil
	}
	return nil, errors.New("Archetype does not exist")
}

// GetArchetypeByName gets the given archetype by string if it exists.
func (m *Manager) GetArchetypeByName(name string) (archetype *Archetype, err error) {
	archID := m.Strings.Acquire(name)
	if _, ok := m.archetypes[archID]; ok {
		return m.archetypes[archID], nil
	}
	return nil, errors.New("Archetype does not exist")
}

func (m *Manager) parseAnimationFiles() error {
	l := log.WithFields(log.Fields{
		"path": m.archetypesPath,
	})
	l.Print("Animations: Loading...")
	err := filepath.Walk(m.archetypesPath, func(filepath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if strings.HasSuffix(filepath, ".anim.yaml") {
				err = m.parseAnimationFile(filepath)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		l.Error("Error walking the path", err)
	}

	l.WithFields(log.Fields{
		"count": len(m.animations),
	}).Print("Animations: Done!")
	return nil
}

// parseAnimationFile parses the given animation file into our animations field.
func (m *Manager) parseAnimationFile(filepath string) error {
	r, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	animationsMap := make(map[string]AnimationPre)

	if err = yaml.Unmarshal(r, &animationsMap); err != nil {
		return err
	}
	for k, animation := range animationsMap {
		animID := m.Strings.Acquire(k)

		parsedAnimation := &Animation{
			Faces:       make(map[StringID][]AnimationFrame),
			RandomFrame: animation.RandomFrame,
		}

		for faceKey, face := range animation.Faces {
			faceID := m.Strings.Acquire(faceKey)

			parsedAnimation.Faces[faceID] = make([]AnimationFrame, 0)

			for _, frame := range face {
				imageID := m.Strings.Acquire(frame.Image)
				parsedFrame := AnimationFrame{
					ImageID: imageID,
					Time:    frame.Time,
					X:       frame.X,
					Y:       frame.Y,
				}
				parsedAnimation.Faces[faceID] = append(parsedAnimation.Faces[faceID], parsedFrame)
			}
		}
		m.animations[animID] = parsedAnimation
	}
	return nil
}

func (m *Manager) buildImagesMap() error {
	l := log.WithFields(log.Fields{
		"path": m.archetypesPath,
	})
	l.Print("imageFileMap: Loading...")
	err := filepath.Walk(m.archetypesPath, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if path.Ext(fpath) == ".png" {
				shortpath := filepath.ToSlash(fpath[len(m.archetypesPath)+1:])
				id := m.Strings.Acquire(shortpath)
				_, err := m.imageFileMap.BuildCRC(id, fpath)

				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		l.Error("Error walking the path", err)
		return err
	}
	l.WithFields(log.Fields{
		"count": len(m.imageFileMap.Paths),
	}).Print("imageFileMap: Done!")
	return nil
}

func (m *Manager) parseAudioFiles() error {
	l := log.WithFields(log.Fields{
		"path": m.audioPath,
	})
	l.Print("Audio: Loading...")
	err := filepath.Walk(m.audioPath, func(filepath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if strings.HasSuffix(filepath, ".audio.yaml") {
				err = m.parseAudioFile(filepath)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		l.Error("Error walking the path", err)
	}

	l.WithFields(log.Fields{
		"count": len(m.audio),
	}).Print("Audio: Done!")
	return nil
}

// parseAudioFile parses the given audio file into our audio field.
func (m *Manager) parseAudioFile(filepath string) error {
	r, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	audioMap := make(map[string]AudioPre)

	if err = yaml.Unmarshal(r, &audioMap); err != nil {
		return err
	}
	for k, audio := range audioMap {
		audioID := m.Strings.Acquire(k)

		parsedAudio := &Audio{
			SoundSets: make(map[StringID][]AudioSound),
		}

		for soundSetKey, soundSet := range audio.SoundSets {
			soundSetID := m.Strings.Acquire(soundSetKey)

			parsedAudio.SoundSets[soundSetID] = make([]AudioSound, 0)

			for _, sound := range soundSet {
				soundID := m.Strings.Acquire(sound.File)
				parsedSubSound := AudioSound{
					SoundID: soundID,
					Text:    sound.Text,
				}
				parsedAudio.SoundSets[soundSetID] = append(parsedAudio.SoundSets[soundSetID], parsedSubSound)
			}
		}
		m.audio[audioID] = parsedAudio
	}
	return nil
}

func (m *Manager) buildSoundsMap() error {
	l := log.WithFields(log.Fields{
		"path": m.audioPath,
	})
	l.Print("soundFileMap: Loading...")
	err := filepath.Walk(m.audioPath, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if path.Ext(fpath) == ".flac" || path.Ext(fpath) == ".ogg" {
				shortpath := filepath.ToSlash(fpath[len(m.audioPath)+1:])
				id := m.Strings.Acquire(shortpath)
				_, err := m.soundFileMap.BuildCRC(id, fpath)

				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		l.Error("Error walking the path", err)
		return err
	}
	l.WithFields(log.Fields{
		"count": len(m.soundFileMap.Paths),
	}).Print("soundFileMap: Done!")
	return nil
}

func (m *Manager) parseMapFile(filepath string) error {
	l := log.WithFields(log.Fields{
		"mapset": filepath,
	})

	r, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	maps := make(map[string]*Map)

	if err = yaml.Unmarshal(r, &maps); err != nil {
		return err
	}
	for k, v := range maps {
		l = l.WithFields(log.Fields{
			"map": k,
		})
		// Acquire our ArchIDs for Tiles
		for y := range v.Tiles {
			for x := range v.Tiles[y] {
				for z := range v.Tiles[y][x] {
					for i := range v.Tiles[y][x][z] {
						l = l.WithFields(log.Fields{
							"y": y,
							"x": x,
							"z": y,
							"i": i,
						})
						// We also process and compile our tiles so as to allow for XTREME custom archs in maps!
						if err := m.ProcessArchetype(&v.Tiles[y][x][z][i]); err != nil {
							l.Warnln(err)
							continue
						}
						if err := m.CompileArchetype(&v.Tiles[y][x][z][i]); err != nil {
							l.Warnln(err)
							continue
						}
					}
				}
			}
		}
		m.maps[k] = v
		m.maps[k].MapID = m.Strings.Acquire(k)
		m.maps[k].DataName = k
		m.maps[k].Filepath = filepath
	}

	return nil
}

func (m *Manager) parseMapFiles() error {
	l := log.WithFields(log.Fields{
		"path": m.mapsPath,
	})

	l.Print("Maps: Loading...")
	err := filepath.Walk(m.mapsPath, func(filepath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if strings.HasSuffix(filepath, ".map.yaml") {
				err = m.parseMapFile(filepath)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		l.Errorln(err)
	}
	l.WithFields(log.Fields{
		"count": len(m.maps),
	}).Println("Maps: Done!")
	return nil
}

// GetMap gets the given map by name if it exists.
func (m *Manager) GetMap(name string) (Map *Map, err error) {
	if _, ok := m.maps[name]; ok {
		return m.maps[name], nil
	}
	return nil, errors.New("Map does not exist")
}

// ReloadMap attempts to reload the given map. This does not restart any running instances.
func (m *Manager) ReloadMap(name string) (err error) {
	gmap, err := m.GetMap(name)
	if err != nil {
		return err
	}
	return m.parseMapFile(gmap.Filepath)
}

// ReloadMapFile attempts to reload the given maps by file. This does not restart any running instances.
func (m *Manager) ReloadMapFile(file string) (err error) {
	return m.parseMapFile(filepath.Join(m.mapsPath, file+".map.yaml"))
}

// Setup sets up the data Manager for use by the server.
func (m *Manager) Setup(config *config.Config) error {
	m.archetypes = make(map[StringID]*Archetype)
	m.animations = make(map[StringID]*Animation)
	m.audio = make(map[StringID]*Audio)
	m.maps = make(map[string]*Map)
	m.loadedUsers = make(map[string]*User)
	m.Strings = NewStrings()
	m.imageFileMap = NewFileMap()
	m.soundFileMap = NewFileMap()
	m.cryptParams = cryptParams{
		memory:      64 * 1024,
		iterations:  12,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	}
	m.TypeHints = make(map[uint32]string)
	m.Slots = make(map[uint32]string)
	// Get the parent dir of command; should resolve like /path/bin/server -> /path/
	dir, err := filepath.Abs(os.Args[0])
	if err != nil {
		log.Fatal(err)
		return nil
	}
	dir = filepath.Dir(filepath.Dir(dir))
	// Data
	dataPath := path.Join(dir, "share", "chimera")
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		log.Fatal(err)
		return err
	}
	m.dataPath = dataPath
	m.archetypesPath = path.Join(dataPath, "archetypes")
	if _, err := os.Stat(m.archetypesPath); os.IsNotExist(err) {
		log.Fatal(err)
		return err
	}
	m.mapsPath = path.Join(dataPath, "maps")
	if _, err := os.Stat(m.mapsPath); os.IsNotExist(err) {
		log.Fatal(err)
		return err
	}
	m.audioPath = path.Join(dataPath, "audio")
	if _, err := os.Stat(m.audioPath); os.IsNotExist(err) {
		log.Fatal(err)
		return err
	}
	// Variable Data
	varPath := path.Join(dir, "var", "chimera")
	if _, err := os.Stat(varPath); os.IsNotExist(err) {
		if err = os.MkdirAll(varPath, os.ModePerm); err != nil {
			log.Fatal(err)
			return err
		}
	}
	m.varPath = varPath
	m.usersPath = path.Join(varPath, "players")
	if _, err := os.Stat(m.usersPath); os.IsNotExist(err) {
		if err = os.Mkdir(m.usersPath, os.ModePerm); err != nil {
			log.Fatal(err)
			return err
		}
	}
	// Etc Data
	etcPath := path.Join(dir, "etc", "chimera")
	if _, err := os.Stat(etcPath); os.IsNotExist(err) {
		if err = os.MkdirAll(etcPath, os.ModePerm); err != nil {
			log.Fatal(err)
			return err
		}
	}
	m.etcPath = etcPath
	/*
	  m.musicPath = path.Join(dataPath, "music")
	  if _, err := os.Stat(m.musicPath); os.IsNotExist(err) {
	    log.Fatal(err)
	    return err
	  }
	  m.soundPath = path.Join(dataPath, "sounds")
	  if _, err := os.Stat(m.soundPath); os.IsNotExist(err) {
	    log.Fatal(err)
	    return err
	  }*/
	// Images
	err = m.buildImagesMap()
	if err != nil {
		log.Fatal(err)
		return err
	}
	// Sounds
	err = m.buildSoundsMap()
	if err != nil {
		log.Fatal(err)
		return err
	}
	// Animations
	// Read animations config
	animationsConfigPath := path.Join(m.archetypesPath, "config.yaml")
	r, err := ioutil.ReadFile(animationsConfigPath)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(r, &m.AnimationsConfig); err != nil {
		return err
	}
	// Read animation files
	err = m.parseAnimationFiles()
	if err != nil {
		log.Fatal(err)
		return err
	}
	// Read audio files
	err = m.parseAudioFiles()
	if err != nil {
		log.Fatal(err)
		return err
	}
	// Archetypes
	err = m.parseArchetypeFiles()
	if err != nil {
		log.Fatal(err)
		return err
	}
	// Maps!
	err = m.parseMapFiles()
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

// GetEtcPath returns the path to the current etc directory.
func (m *Manager) GetEtcPath() string {
	return m.etcPath
}

// GetPCArchetypes returns the underlying *Archetype slice for player character archetypes.
func (m *Manager) GetPCArchetypes() []*Archetype {
	return m.pcArchetypes
}

// GetGeneraArchetypes returns the underlying *Archetype slice for genera archetypes.
func (m *Manager) GetGeneraArchetypes() []*Archetype {
	return m.generaArchetypes
}

// GetSpeciesArchetypes returns the underlying *Archetype slice for species archetypes.
func (m *Manager) GetSpeciesArchetypes() []*Archetype {
	return m.speciesArchetypes
}

// GetFactionArchetypes returns the underlying *Archetype slice for faction archetypes.
func (m *Manager) GetFactionArchetypes() []*Archetype {
	return m.factionArchetypes
}

// GetString returns the StringId associated with the passed name.
func (m *Manager) GetString(name string) StringID {
	return m.Strings.Acquire(name)
}

// GetAnimation returns a pointer to an Animation that corresponds to the passed animation ID. Returns nil if none is found.
func (m *Manager) GetAnimation(animID StringID) (*Animation, error) {
	if anim, ok := m.animations[animID]; ok {
		return anim, nil
	}
	return nil, errors.New("Animation does not exist")
}

// GetAnimationByName returns a pointer to an Animation that corresponds to the passed animation name. Returns nil if none is found.
func (m *Manager) GetAnimationByName(name string) (*Animation, error) {
	return m.GetAnimation(m.Strings.Acquire(name))
}

// GetAnimationFrame returns the AnimationFrame for an animation ID, its face ID, and an entry index.
func (m *Manager) GetAnimationFrame(animID StringID, faceID StringID, index int) AnimationFrame {
	if anim, ok := m.animations[animID]; ok {
		if face, ok := anim.Faces[faceID]; ok {
			if index >= 0 && index < len(face) {
				return face[index]
			}
		}
	}
	return AnimationFrame{0, 0, 0, 0}
}

// GetImageData returns the image bytes associated with the provided image ID.
func (m *Manager) GetImageData(imageID StringID) ([]byte, error) {
	return m.imageFileMap.GetBytes(imageID)
}

// GetAudio returns a pointer to an Audio that corresponds to the passed Audio ID. Returns nil if none is found.
func (m *Manager) GetAudio(audioID StringID) (*Audio, error) {
	if _, ok := m.audio[audioID]; ok {
		return m.audio[audioID], nil
	}
	return nil, errors.New("Audio does not exist")
}

// GetSoundData returns the sound bytes associated with the provided sound ID.
func (m *Manager) GetSoundData(soundID StringID) ([]byte, error) {
	return m.soundFileMap.GetBytes(soundID)
}

/*func (m *Manager) createObject(which string) World.GameObject {
  if val, ok := m.templates[string]; ok {
  }
  return new(World.GameObject)
}*/
