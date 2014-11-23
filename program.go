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
}

type ExecutionResult struct {
	Messages   chan *ExecutionMessage
	ExitStatus chan ExitCode
}

type ExecutionMessage struct {
	ProgramName string `json:"programName"`
	MessageType string `json:"messageType"`
	Line        string `json:"line"`
}

func forwardOutput(p *Program, messageType string, r io.Reader,
	output chan *ExecutionMessage, finished chan interface{}) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		s := scanner.Text()
		log.Println(p.Name, messageType, s)
		output <- &ExecutionMessage{p.Name, messageType, s + "\n"}
	}

	if err := scanner.Err(); err != nil {
		log.Println(p.Name, "scanner error", err)
	}

	finished <- struct{}{}
}

func (p *Program) Execute() (*ExecutionResult, error) {

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
	exit := make(chan ExitCode)
	result := &ExecutionResult{messages, exit}
	stdoutFinished := make(chan interface{})
	stderrFinished := make(chan interface{})

	go forwardOutput(p, "out", stdout, messages, stdoutFinished)
	go forwardOutput(p, "err", stderr, messages, stderrFinished)

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
			log.Println(p.Name, "successfully completed")
			messages <- &ExecutionMessage{p.Name, "ok", "successfully completed"}
			// missing ExitCode in this case?
			return
		}

		exitError := err.(*exec.ExitError)
		waitStatus := exitError.Sys().(syscall.WaitStatus)
		exitCode := waitStatus.ExitStatus()
		log.Println(p.Name, "exited with status", exitCode)

		messages <- &ExecutionMessage{p.Name, "fail", fmt.Sprintln("exited with status", exitCode)}
		exit <- ExitCode(exitCode)
	}()

	return result, nil
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
			log.Println("program executable:", commandPath)

			programs = append(programs, &Program{info.Name(), commandPath})
		}
	}

	return programs, nil
}
