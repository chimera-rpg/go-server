package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/chimera-rpg/go-server/config"
	"github.com/chimera-rpg/go-server/server"

	"gopkg.in/yaml.v2"
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.Print("Starting Chimera (golang)")
	// Copied from data/Manager.go
	// Get the parent dir of command; should resolve like /path/bin/server -> /path/
	dir, err := filepath.Abs(os.Args[0])
	if err != nil {
		log.Fatal(err)
	}
	dir = filepath.Dir(filepath.Dir(dir))
	// Get our default configuration path.
	cfgPath := path.Join(dir, "etc", "chimera", "config.yml")
	noPrompt := false

	// Load up flags.
	flag.StringVar(&cfgPath, "config", cfgPath, "configuration file")
	flag.StringVar(&cfgPath, "c", cfgPath, "configuration file (shorthand)")
	flag.BoolVar(&noPrompt, "no-prompt", noPrompt, "Disable command prompt")
	flag.Parse()

	// Setup our default configuration.
	cfg := config.Config{
		Address:  ":1337",
		UseTLS:   true,
		TLSKey:   "server.key",
		TLSCert:  "server.crt",
		Tickrate: 16,
	}
	// Load in our configuration.
	log.Printf("Attempting to load config from \"%s\"\n", cfgPath)
	r, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		// Ensure path to cfg exists.
		if _, err := os.Stat(filepath.Dir(cfgPath)); os.IsNotExist(err) {
			if err = os.MkdirAll(filepath.Dir(cfgPath), os.ModePerm); err != nil {
				log.Fatal(err)
			}
		}
		// Write out default config.
		log.Printf("Creating default config \"%s\"\n", cfgPath)
		bytes, _ := yaml.Marshal(&cfg)
		err = ioutil.WriteFile(cfgPath, bytes, 0644)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		if err = yaml.Unmarshal(r, &cfg); err != nil {
			log.Fatal(err)
		}
	}

	// Begin listening on all interfaces.
	s := server.New()
	if err := s.Setup(&cfg); err != nil {
		log.Fatal(err)
	}

	// Start our server either securely or insecurely.
	if cfg.UseTLS == true {
		if err := s.SecureStart(); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := s.Start(); err != nil {
			log.Fatal(err)
		}
	}

	// Main co-processing looperino
	log.Printf("Ticking at %dms\n", cfg.Tickrate)
	ticker := time.NewTicker(time.Millisecond * time.Duration(cfg.Tickrate))
	go func() {
		lastTime := time.Now()
		for currentTime := range ticker.C {
			timeSinceLastFrame := currentTime.Sub(lastTime)

			s.Update(currentTime, timeSinceLastFrame)

			currentEnd := time.Now()
			//current_elapsed := currentEnd.Sub(currentTime)

			lastTime = currentEnd
		}
	}()
	// Create and initialize our prompt.
	if !noPrompt {
		var prompt Prompt
		prompt.Init(s)
		fmt.Println("Entering prompt. Issue \"help\" for commands.")
		prompt.Capture()
		go prompt.ShowPrompt()
	}
	<-s.End
}
