package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

type Program struct {
	Name        string
	CommandPath string
}

const BUFFER_SIZE = 1000

type ExitCode int
const (
	Success   = 0
	Retryable = 1
	Failed    = 2
)

type ExecutionResult struct {
	Messages chan string
	ExitStatus chan ExitCode
}

func (p *Program) Execute() (*ExecutionResult, error) {
	log.Println("executing", p.CommandPath)
	cmd := exec.Command(p.CommandPath)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	messages := make(chan string, BUFFER_SIZE)
	exit := make(chan ExitCode)
	
	result := &ExecutionResult{messages, exit}

	go func() {
		log.Println(p.Name, "waiting to complete")
		err := cmd.Wait()
		if err == nil {
			log.Println(p.Name, "successfully completed")
			result.Messages<-fmt.Sprintln("successfully completed")
			return
		}
		
		exitError := err.(*exec.ExitError)
		waitStatus := exitError.Sys().(syscall.WaitStatus)
		exitCode := waitStatus.ExitStatus()
		log.Println(p.Name, "exited with status", exitCode)
		
		result.Messages<-fmt.Sprintln("exited with status", exitCode)
		result.ExitStatus<-ExitCode(exitCode)
	}()

	go func() {
		scanner := bufio.NewScanner(stdout)

		for scanner.Scan() {
			s := scanner.Text()
			log.Println(p.Name, s)
			result.Messages <- s
		}

		if err := scanner.Err(); err != nil {
			log.Println(p.Name, "scanner error", err)
		}
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
