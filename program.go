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

const (
	Success   = 0
	Retryable = 1
	Failed    = 2
)

func (p *Program) Execute() (chan string, error) {
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

	go func() {
		log.Println(p.Name, "waiting to complete")
		err := cmd.Wait()
		if err == nil {
			log.Println(p.Name, "successfully completed")
			messages<-fmt.Sprintln("successfully completed")
			return
		}
		
		exitError := err.(*exec.ExitError)
		waitStatus := exitError.Sys().(syscall.WaitStatus)
		exitCode := waitStatus.ExitStatus()
		log.Println(p.Name, "exited with status", exitCode)
		
		messages<-fmt.Sprintln("exited with status", exitCode)
	}()

	go func() {
		scanner := bufio.NewScanner(stdout)

		for scanner.Scan() {
			s := scanner.Text()
			log.Println(p.Name, s)
			messages <- s
		}

		if err := scanner.Err(); err != nil {
			log.Println(p.Name, "scanner error", err)
		}
	}()

	return messages, nil
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
