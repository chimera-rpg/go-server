package main

import (
	"log"
	/*  "bufio"
	    "strings"*/
	"time"
	//
	"github.com/chimera-rpg/go-server/server"
)

func main() {
	log.Print("Starting Chimera (golang)")

	// Begin listening on all interfaces.
	s := server.New()
	if err := s.Setup(); err != nil {
		log.Print(err)
	}
	// TODO: We need to pass in a server YAML file that contains information such as whether to use TLS or not. Ideally we would also merge this with passed flags. For now we'll try to start securely, and if that fails, we do an insecure start. This presumes that server.crt exists in the CWD.
	if err := s.SecureStart(); err != nil {
		log.Print(err)
		if err := s.Start(); err != nil {
			log.Print(err)
			return
		}
	}

	// Main co-processing looperino
	ticker := time.NewTicker(time.Millisecond * 100)
	go func() {
		lastTime := time.Now()
		for currentTime := range ticker.C {
			timeSinceLastFrame := currentTime.Sub(lastTime)

			s.Update(int64(timeSinceLastFrame) / 100000)

			currentEnd := time.Now()
			//current_elapsed := currentEnd.Sub(currentTime)

			lastTime = currentEnd
		}
	}()
	<-s.End
}
