package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
)

const BUFFER_SIZE = 1000

const (
	Success   = 0
	Retryable = 1
	Failed    = 2
)

type ExitCode int

type Program struct {
	Name        string
	CommandPath string
	MainSource  string
	executions  []*Execution
	sync.RWMutex
}

func forwardOutput(execution *Execution, messageType string, r io.Reader, finished chan interface{}) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		execution.SendMessage(messageType, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Println(execution.Program.Name, "scanner error", err)
	}

	finished <- struct{}{}
}

func (p *Program) Execute() (*Execution, error) {
	p.Lock()
	defer p.Unlock()

	log.Println("executing", p.CommandPath)
	cmd := exec.Command(p.CommandPath)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	messages := make(chan *ExecutionMessage, BUFFER_SIZE)
	execution := NewExecution(p, messages)
	stdoutFinished := make(chan interface{})
	stderrFinished := make(chan interface{})

	go forwardOutput(execution, "out", stdout, stdoutFinished)
	go forwardOutput(execution, "err", stderr, stderrFinished)

	go func() {
		defer close(messages)
		defer close(stdoutFinished)
		defer close(stderrFinished)

		log.Println(p.Name, "waiting to for stdout and stderr to be read")

		// docs say we shouldn't call cmd.Wait() until all has been read, hence
		// the need for the 'finished' channels
		<-stdoutFinished
		<-stderrFinished

		err := cmd.Wait()
		if err == nil {
			execution.SendMessage("ok", "successfully completed")
			execution.Finish(ExitCode(0))
			return
		}

		exitCode, err := extractExitCode(err)

		if err != nil {
			log.Println(p.Name, "failed to run", err)
			execution.SendMessage("fail", fmt.Sprint("failed to run ", err))
			execution.Finish(ExitCode(-1)) //?
			return
		}

		log.Println(p.Name, "exited with status", exitCode)
		execution.SendMessage("fail", fmt.Sprint("exited with status ", exitCode))
		execution.Finish(exitCode)
	}()

	p.executions = append(p.executions, execution)

	return execution, nil
}

func (p *Program) Executions() []*Execution {
	p.RLock()
	defer p.RUnlock()
	return p.executions
}

func extractExitCode(err error) (ExitCode, error) {
	switch ex := err.(type) {
	case *exec.ExitError:
		return ExitCode(ex.Sys().(syscall.WaitStatus).ExitStatus()), nil // assume Unix
	default:
		return 0, err
	}
}

func readDir(dir string) ([]*Program, error) {
	log.Println("looking for programs in", dir)
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	programs := []*Program{}

	for _, info := range infos {
		commandPath := filepath.Join(dir, info.Name(), "main")
		_, err := os.Stat(commandPath)

		if err == nil {
			mainSource, err := ioutil.ReadFile(commandPath)

			if err == nil {
				log.Println("program executable:", commandPath)
				programs = append(programs, &Program{
					Name:        info.Name(),
					CommandPath: commandPath,
					MainSource:  string(mainSource),
				})
			}
		}
	}

	return programs, nil
}
