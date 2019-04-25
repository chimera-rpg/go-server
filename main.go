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
	s := server.New(":1337")
	s.Start()

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
