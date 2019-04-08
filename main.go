package main

import (
  "log"
/*  "bufio"
  "strings"*/
  "time"
  //
  "github.com/chimera-rpg/go-server/GameServer"
)

func main() {
  log.Print("Starting Chimera (golang)")

  // Begin listening on all interfaces.
  server := GameServer.New(":1337")
  server.Start()

  // Main co-processing looperino
  ticker := time.NewTicker(time.Millisecond * 100)
  go func() {
    last_time := time.Now()
    for current_time := range ticker.C {
      time_since_last_frame := current_time.Sub(last_time)

      server.Update(int64(time_since_last_frame)/100000)

      current_end := time.Now()
      //current_elapsed := current_end.Sub(current_time)

      last_time = current_end
    }
  }()
  <-server.End
}
