package data

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"

	"github.com/chimera-rpg/go-server/config"
)

// Manager is the controlling type for accessing archetypes, maps, player
// files and any other data.
type Manager struct {
	dataPath       string
	archetypesPath string
	varPath        string
	etcPath        string
	usersPath      string
	usersMutex     sync.Mutex
	//musicPath string
	//soundPath string
	mapsPath   string
	archetypes map[StringID]*Archetype // Full Map of archetypes.
	//animations map[string]*Animation // Full Map of animations.
	animations map[StringID]*Animation // ID to Animation map
	// Hmm... almost map[uint32]*Archetype... with CRC id
	Strings      StringsMap
	imageFileMap FileMap
	// images map[string][]bytes
	generaArchetypes  []*Archetype     // Slice of genera archetypes.
	speciesArchetypes []*Archetype     // Slice of species archetypes.
	pcArchetypes      []*Archetype     // Player Character archetypes, used for creating new characters.
	maps              map[string]*Map  // Full map of Maps.
	loadedUsers       map[string]*User // Map of loaded Players
	cryptParams       cryptParams      // Cryptography parameters
}

type objectTemplate struct {
	variables map[string]objectVariable
}

type objectVariable struct {
	key   string
	value string
}

func (m *Manager) parseArchetypeFile(filepath string) error {
	r, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	archetypesMap := make(map[string]Archetype)

	if err = yaml.Unmarshal(r, &archetypesMap); err != nil {
		return err
	}
	for k, archetype := range archetypesMap {
		archID := m.Strings.Acquire(k)
		m.archetypes[archID] = &archetype
		m.archetypes[archID].SelfID = archID
	}
	return nil
}

func (m *Manager) postProcessArchetype(archetype *Archetype) error {
	if archetype.Arch != "" {
		inheritArch, err := m.GetArchetypeByName(archetype.Arch)
		if err != nil {
			return err
		}
		archetype.InheritArch = inheritArch
	}
	if archetype.Anim != "" {
		archetype.AnimID = m.Strings.Acquire(archetype.Anim)
		archetype.Anim = ""
	}
	//archetype.Arch = archName
	//archetype.ArchID = m.Strings.Acquire(archName)
	for i := range archetype.Inventory {
		if err := m.postProcessArchetype(&archetype.Inventory[i]); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) parseArchetypeFiles() error {
	log.Print("Archetypes: Loading...")
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
		log.Printf("Error walking the path %s: %v\n", m.archetypesPath, err)
	}
	// Post-process our archetypes so they properly set up inheritance relationships.
	for _, archetype := range m.archetypes {
		m.postProcessArchetype(archetype)
	}
	fmt.Printf("%+v\n", m.archetypes)

	log.Printf("%d archetypes loaded.\n", len(m.archetypes))
	log.Printf("%d Genera.", m.buildGeneraArchetypes())
	log.Printf("%d Species.", m.buildSpeciesArchetypes())
	log.Printf("%d PCs.", m.buildPCArchetypes())
	for _, v := range m.pcArchetypes {
		name, _ := v.Name.Get()
		log.Printf("%s ", name)
	}
	log.Print("Archetypes: Done!")
	return nil
}

func (m *Manager) buildPCArchetypes() int {
	oldCount := len(m.pcArchetypes)
	for _, v := range m.archetypes {
		if v.Type == ArchetypePC {
			m.pcArchetypes = append(m.pcArchetypes, v)
		}
	}
	return len(m.pcArchetypes) - oldCount
}

func (m *Manager) buildGeneraArchetypes() int {
	oldCount := len(m.generaArchetypes)
	for _, v := range m.archetypes {
		if v.Type == ArchetypeGenus {
			m.generaArchetypes = append(m.generaArchetypes, v)
		}
	}
	return len(m.generaArchetypes) - oldCount
}

func (m *Manager) buildSpeciesArchetypes() int {
	oldCount := len(m.speciesArchetypes)
	for _, v := range m.archetypes {
		if v.Type == ArchetypeSpecies {
			m.speciesArchetypes = append(m.speciesArchetypes, v)
		}
	}
	return len(m.speciesArchetypes) - oldCount
}

// GetArchetype gets the given archetype by id if it exists.
func (m *Manager) GetArchetype(archID FileID) (archetype *Archetype, err error) {
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
	log.Print("Animations: Loading...")
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
		log.Printf("Error walking the path %s: %v\n", m.archetypesPath, err)
	}
	log.Printf("%d animations loaded.\n", len(m.animations))

	log.Print("Animations: Done!")
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
			Faces: make(map[StringID][]AnimationFrame),
		}

		for faceKey, face := range animation.Faces {
			faceID := m.Strings.Acquire(faceKey)

			parsedAnimation.Faces[faceID] = make([]AnimationFrame, 0)

			for _, frame := range face {
				imageID := m.Strings.Acquire(frame.Image)
				parsedFrame := AnimationFrame{
					ImageID: imageID,
					Time:    frame.Time,
				}
				parsedAnimation.Faces[faceID] = append(parsedAnimation.Faces[faceID], parsedFrame)
			}
		}
		m.animations[animID] = parsedAnimation
	}
	return nil
}

func (m *Manager) buildImagesMap() error {
	log.Print("imageFileMap: Loading...")
	err := filepath.Walk(m.archetypesPath, func(filepath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if path.Ext(filepath) == ".png" {
				shortpath := filepath[len(m.archetypesPath)+1:]
				id := m.Strings.Acquire(shortpath)
				crc, err := m.imageFileMap.BuildCRC(id, filepath)
				log.Printf("%s(%d) = %d\n", shortpath, id, crc)

				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("Error walking the path %s: %v\n", m.archetypesPath, err)
		return err
	}
	log.Print("imageFileMap: Done!")
	return nil
}

func (m *Manager) parseMapFile(filepath string) error {
	r, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	maps := make(map[string]Map)

	if err = yaml.Unmarshal(r, &maps); err != nil {
		return err
	}
	log.Printf("%+v\n", maps)
	for k, v := range maps {
		// Acquire our ArchIDs for Tiles
		for y := range v.Tiles {
			for x := range v.Tiles[y] {
				for z := range v.Tiles[y][x] {
					for i := range v.Tiles[y][x][z] {
						// We also post-process our tiles so as to allow for XTREME custom archs in maps!
						if err := m.postProcessArchetype(&v.Tiles[y][x][z][i]); err != nil {
							return err
						} else {
							log.Printf("%+v\n", v.Tiles[y][x][z][i])
						}
					}
				}
			}
		}
		m.maps[k] = &v
		log.Printf("%+v\n", v)
	}

	return nil
}

func (m *Manager) parseMapFiles() error {
	log.Printf("Maps: Loading from \"%s\"...\n", m.mapsPath)
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
		log.Printf("Error walking the path %s: %v\n", m.mapsPath, err)
	}
	log.Printf("%d maps loaded.", len(m.maps))
	log.Print("Maps: Done!")
	return nil
}

// GetMap gets the given map by name if it exists.
func (m *Manager) GetMap(name string) (Map *Map, err error) {
	if _, ok := m.maps[name]; ok {
		return m.maps[name], nil
	}
	return nil, errors.New("Map does not exist")
}

// Setup sets up the data Manager for use by the server.
func (m *Manager) Setup(config *config.Config) error {
	m.archetypes = make(map[StringID]*Archetype)
	m.animations = make(map[StringID]*Animation)
	m.maps = make(map[string]*Map)
	m.loadedUsers = make(map[string]*User)
	m.Strings = NewStringsMap()
	m.imageFileMap = NewFileMap()
	m.cryptParams = cryptParams{
		memory:      64 * 1024,
		iterations:  12,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	}
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
	// Now we can load our archetypes into memory.
	err = m.parseArchetypeFiles()
	if err != nil {
		log.Fatal(err)
		return err
	}
	// Images
	err = m.buildImagesMap()
	if err != nil {
		log.Fatal(err)
		return err
	}
	// Animations
	err = m.parseAnimationFiles()
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

// GetString returns the StringId associated with the passed name.
func (m *Manager) GetString(name string) StringID {
	return m.Strings.Acquire(name)
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
	return AnimationFrame{0, 0}
}

/*func (m *Manager) createObject(which string) World.GameObject {
  if val, ok := m.templates[string]; ok {
  }
  return new(World.GameObject)
}*/
