package main

import (
	"bufio"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
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
}

func (p *Prompt) Init() (err error) {
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

func (p *Prompt) handleCommand(c string, args ...string) error {
	if c == "log" {
		p.Uncapture()
		fmt.Println("Outputting logs. Press Enter to re-open console.")
		p.inputScanner.Scan()
		p.Capture()
		p.ShowPrompt()
	} else if c == "help" {
		fmt.Fprintf(p.stdout, "\"log\" to show log output, \"quit\" to quit\n")
		p.ShowPrompt()
	} else if c == "quit" {
		os.Exit(0)
	} else if c == "" {
		p.ShowPrompt()
	} else {
		return errors.New(fmt.Sprintf("unknown command \"%s\"", c))
	}
	return nil
}
