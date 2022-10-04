package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/chimera-rpg/go-server/server"
	log "github.com/sirupsen/logrus"
)

type Prompt struct {
	sync.Mutex
	inputScanner *bufio.Scanner
	stdout       *os.File
	stderr       *os.File
	stdoutWriter *os.File
	stdoutReader *os.File
	stderrWriter *os.File
	stderrReader *os.File
	logOutput    io.Writer
	logWriter    *os.File
	logReader    *os.File
	gameServer   *server.GameServer
}

func (p *Prompt) Init(s *server.GameServer) (err error) {
	p.gameServer = s
	p.stdout = os.Stdout
	p.stderr = os.Stderr
	p.inputScanner = bufio.NewScanner(os.Stdin)
	err = p.makeRWs()
	return nil
}

func (p *Prompt) makeRWs() (err error) {
	p.stdoutReader, p.stdoutWriter, err = os.Pipe()
	if err != nil {
		return err
	}
	p.stderrReader, p.stderrWriter, err = os.Pipe()
	if err != nil {
		return err
	}
	p.logReader, p.logWriter, err = os.Pipe()
	if err != nil {
		return err
	}
	return nil
}

func (p *Prompt) Capture() {
	p.Lock()
	defer p.Unlock()
	// Recreate pairs... this is dumb.
	p.makeRWs()
	// Store os and log destinations
	p.stdout = os.Stdout
	p.stderr = os.Stderr
	// p.logOutput = log.StandardLogger().Writer()
	// Replace os and log output
	os.Stdout = p.stdoutWriter
	os.Stderr = p.stderrWriter
	log.SetOutput(p.stdoutWriter)
}

func (p *Prompt) Uncapture() {
	p.Lock()
	defer p.Unlock()
	// Restore os and log destinations
	os.Stdout = p.stdout
	os.Stderr = p.stderr
	log.SetOutput(p.stdout)

	// Copy any held data to stdout/stderr.
	outC := make(chan struct{})
	go func() {
		io.Copy(p.stdout, p.stdoutReader)
		io.Copy(p.stderr, p.stderrReader)
		io.Copy(log.StandardLogger().Writer(), p.logReader)
		outC <- struct{}{}
	}()
	p.stdoutWriter.Close()
	p.stderrWriter.Close()
	p.logWriter.Close()
	<-outC
}

func (p *Prompt) ShowPrompt() {
	fmt.Fprintf(p.stdout, "> ")
	p.inputScanner.Scan()
	if err := p.handleCommand(p.inputScanner.Text()); err != nil {
		fmt.Fprintln(p.stdout, err)
		p.ShowPrompt()
	}
}

func (p *Prompt) handleCommand(c string) error {
	r := csv.NewReader(strings.NewReader(c))
	r.Comma = ' '
	args, err := r.Read()
	if err != nil {
		return err
	}
	if len(args) == 0 {
		p.ShowPrompt()
		return nil
	}

	if args[0] == "log" {
		p.Uncapture()
		fmt.Println("Outputting logs. Press Enter to re-open console.")
		p.inputScanner.Scan()
		p.Capture()
		p.ShowPrompt()
	} else if args[0] == "help" {
		fmt.Fprintf(p.stdout, "\tlog\tshow log output\n\tplayers\tlist players\n\tlookup\tlookup information\n\tmap\treload or restart maps\n\tquit\tshutdown and close\n")
		p.ShowPrompt()
	} else if args[0] == "lookup" {
		if len(args) != 3 {
			fmt.Fprint(p.stdout, "Usage:\n\tlookup string <stringID>\n\tlookup map \"<name>\"\n\tlookup object <objectID>\n\tlookup archetype <stringID>|\"<archetype name>\"\n\tlookup animation <stringID>|\"<animation name>\"\n\tlookup player <objectID|\"<username>\"\n")
		} else {
			if args[1] == "string" {
				u, err := strconv.ParseUint(args[2], 10, 32)
				if err != nil {
					fmt.Fprintf(p.stdout, err.Error())
				} else {
					str := p.gameServer.GetDataManager().Strings.Lookup(uint32(u))
					fmt.Fprintf(p.stdout, "%d => \"%s\"\n", uint32(u), str)
				}
			} else if args[1] == "map" {
				m := p.gameServer.GetWorld().GetMap(args[2])
				fmt.Fprintf(p.stdout, "%+v\n", m)
			} else if args[1] == "object" {
				u, err := strconv.ParseUint(args[2], 10, 32)
				if err != nil {
					fmt.Fprintf(p.stdout, err.Error())
				} else {
					o := p.gameServer.GetWorld().GetObject(uint32(u))
					fmt.Fprintf(p.stdout, "%d => %+v\n", u, o)
				}
			} else if args[1] == "archetype" {
				u, err := strconv.ParseUint(args[2], 10, 32)
				if err != nil {
					arch, _ := p.gameServer.GetDataManager().GetArchetypeByName(args[2])
					fmt.Fprintf(p.stdout, "\"%s\" => %+v\n", args[2], arch)
				} else {
					arch, _ := p.gameServer.GetDataManager().GetArchetype(uint32(u))
					fmt.Fprintf(p.stdout, "%d => %+v\n", u, arch)
				}
			} else if args[1] == "animation" {
				u, err := strconv.ParseUint(args[2], 10, 32)
				if err != nil {
					anim, _ := p.gameServer.GetDataManager().GetAnimationByName(args[2])
					fmt.Fprintf(p.stdout, "\"%s\" => %+v\n", args[2], anim)
				} else {
					anim, _ := p.gameServer.GetDataManager().GetAnimation(uint32(u))
					fmt.Fprintf(p.stdout, "%d => %+v\n", u, anim)
				}
			} else if args[1] == "player" {
				u, err := strconv.ParseUint(args[2], 10, 32)
				if err != nil {
					player := p.gameServer.GetWorld().GetPlayerByUsername(args[2])
					fmt.Fprintf(p.stdout, "%s => %+v\n", args[2], player)
				} else {
					player := p.gameServer.GetWorld().GetPlayerByObjectID(uint32(u))
					fmt.Fprintf(p.stdout, "%d => %+v\n", uint32(u), player)
				}
			}
		}
		p.ShowPrompt()
	} else if args[0] == "players" {
		fmt.Fprintf(p.stdout, "%+v\n", p.gameServer.GetWorld().GetPlayers())
		p.ShowPrompt()
	} else if args[0] == "map" {
		if len(args) != 3 {
			fmt.Fprint(p.stdout, "Usage:\n\tmap reloadFile \"<map file>\"\n\tmap reload \"<name>\"\n\tmap restart \"<name>\"\n")
		} else {
			if args[1] == "reload" {
				// TODO: Reload from disque.
				if err := p.gameServer.GetDataManager().ReloadMap(args[2]); err != nil {
					fmt.Fprint(p.stderr, err)
				} else {
					fmt.Fprint(p.stdout, "reloaded")
				}
			} else if args[1] == "reloadFile" {
				if err := p.gameServer.GetDataManager().ReloadMapFile(args[2]); err != nil {
					fmt.Fprint(p.stderr, err)
				} else {
					fmt.Fprint(p.stdout, "reloaded")
				}
			} else if args[2] == "restart" {
				p.gameServer.GetWorld().RestartMap(args[2])
			}
		}
		p.ShowPrompt()
	} else if args[0] == "clock" {
		h, m, s := p.gameServer.GetWorld().Time.Clock()
		fmt.Fprintf(p.stdout, "%02d:%02d:%02d\n", h, m, s)
		p.ShowPrompt()
	} else if args[0] == "date" {
		y, m, d := p.gameServer.GetWorld().Time.Date()
		fmt.Fprintf(p.stdout, "%02d/%02d/%02d, %s %s cycle(%f) of the %d day in the season of %s\n", y, m, d, p.gameServer.GetWorld().Time.Cycle(), p.gameServer.GetWorld().Time.Cycle().Diel(), p.gameServer.GetWorld().Time.Cycle(), d, p.gameServer.GetWorld().Time.Season())
		p.ShowPrompt()
	} else if args[0] == "quit" {
		os.Exit(0)
	} else {
		return errors.New(fmt.Sprintf("unknown command \"%s\"", args[0]))
	}
	return nil
}
