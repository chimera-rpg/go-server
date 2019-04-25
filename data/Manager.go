package data

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

type Manager struct {
	dataPath       string
	archetypesPath string
	//musicPath string
	//soundPath string
	mapsPath     string
	archetypes   map[string]*Archetype // Full Map of archetypes.
	pcArchetypes []*Archetype          // Player Character archetypes, used for creating new characters.
	maps         map[string]*Map       // Full map of Maps.
	//animations map[string]gameAnimation
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
	parser := new(archetypeParser)
	parser.lexer = NewObjectLexer(string(r))
	// Parse our archetypes and merge with existing.
	for k, v := range parser.parse() {
		log.Printf("%s = %v\n", k, v)
		m.archetypes[k] = &v
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
			if path.Ext(filepath) == ".arch" {
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
	log.Printf("%d archetypes loaded.\n", len(m.archetypes))
	log.Printf("%d PC archetypes loaded: ", m.buildPCArchetypes())
	for _, v := range m.pcArchetypes {
		name, _ := v.Name.GetString()
		log.Printf("%s ", name)
	}
	log.Print("Archetypes: Done!")
	return nil
}

func (m *Manager) buildPCArchetypes() int {
	old_count := len(m.pcArchetypes)
	for _, v := range m.archetypes {
		if v.Type == ArchetypePC {
			m.pcArchetypes = append(m.pcArchetypes, v)
		}
	}
	return len(m.pcArchetypes) - old_count
}

func (m *Manager) GetArchetype(name string) (archetype *Archetype, err error) {
	if _, ok := m.archetypes[name]; ok {
		return m.archetypes[name], nil
	}
	return nil, errors.New("Map does not exist")
}

func (m *Manager) parseMapFile(filepath string) error {
	r, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	parser := new(mapParser)
	parser.lexer = NewObjectLexer(string(r))
	//log.Printf("%+v\n", parser.parse())
	// Parse our maps and merge with existing.
	for k, v := range parser.parse() {
		log.Printf("%s = %v\n", k, v)
		m.maps[k] = &v
	}

	return nil
}

func (m *Manager) parseMapFiles() error {
	log.Print("Maps: Loading...")
	err := filepath.Walk(m.mapsPath, func(filepath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if path.Ext(filepath) == ".map" {
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

func (m *Manager) GetMap(name string) (Map *Map, err error) {
	if _, ok := m.maps[name]; ok {
		return m.maps[name], nil
	}
	return nil, errors.New("Map does not exist")
}

func (m *Manager) Setup() error {
	m.archetypes = make(map[string]*Archetype)
	m.maps = make(map[string]*Map)
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return nil
	}
	data_path := path.Join(dir, "share", "chimera")
	if _, err := os.Stat(data_path); os.IsNotExist(err) {
		log.Fatal(err)
		return err
	}
	m.dataPath = data_path
	m.archetypesPath = path.Join(data_path, "archetypes")
	if _, err := os.Stat(m.archetypesPath); os.IsNotExist(err) {
		log.Fatal(err)
		return err
	}
	m.mapsPath = path.Join(data_path, "maps")
	if _, err := os.Stat(m.mapsPath); os.IsNotExist(err) {
		log.Fatal(err)
		return err
	}
	/*
	  m.musicPath = path.Join(data_path, "music")
	  if _, err := os.Stat(m.musicPath); os.IsNotExist(err) {
	    log.Fatal(err)
	    return err
	  }
	  m.soundPath = path.Join(data_path, "sounds")
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
	// Maps!
	err = m.parseMapFiles()
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

/*func (m *Manager) createObject(which string) World.GameObject {
  if val, ok := m.templates[string]; ok {
  }
  return new(World.GameObject)
}*/
