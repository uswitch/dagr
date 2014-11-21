package main

import (
	"bufio"
	"log"
	"os/exec"
)

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

	// go func() {
	//	log.Println("waiting to finish", p.Name)
	//	cmd.Wait()
	//	log.Println("finished", p.Name)
	// }()

	messages := make(chan string, BUFFER_SIZE)

	go func() {
		scanner := bufio.NewScanner(stdout)

		for scanner.Scan() {
			s := scanner.Text()
			log.Println(p.Name, s)
			messages <- s
		}

		if err := scanner.Err(); err != nil {
			log.Println("scanner error", err)
		}
	}()

	return messages, nil
}
