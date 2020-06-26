package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/chimera-rpg/go-server/server"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
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
		fmt.Fprintf(p.stdout, "\"log\" to show log output, \"quit\" to quit\n")
		p.ShowPrompt()
	} else if args[0] == "lookup" {
		if len(args) != 3 {
			fmt.Fprint(p.stdout, "Usage: lookup string <stringID>\n")
		} else {
			if args[1] == "string" {
				u, err := strconv.ParseUint(args[2], 10, 32)
				if err != nil {
					fmt.Fprintf(p.stdout, err.Error())
				} else {
					str := p.gameServer.GetDataManager().Strings.Lookup(uint32(u))
					fmt.Fprintf(p.stdout, "%d => \"%s\"\n", uint32(u), str)
				}
			}
		}
		p.ShowPrompt()
	} else if args[0] == "dump" {
		if len(args) != 3 {
			fmt.Fprintf(p.stdout, "Usage: dump map \"<map name>\"\n")
		} else {
			if args[1] == "map" {
				m := p.gameServer.GetWorld().GetMap(args[2])
				fmt.Fprintf(p.stdout, "%+v\n", m)
			}
		}
		p.ShowPrompt()
	} else if args[0] == "quit" {
		os.Exit(0)
	} else {
		return errors.New(fmt.Sprintf("unknown command \"%s\"", args[0]))
	}
	return nil
}
