package main

import (
	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type Dagr interface {
	AllPrograms() []*Program
	AddExecution(*Program) *Execution
	FindProgram(string) *Program
	FindExecution(string) *Execution
}

type dagrState struct {
	programs   []*Program
	executions map[string]*Execution
	sync.RWMutex
}

func (this *dagrState) FindProgram(name string) *Program {
	this.RLock()
	defer this.RUnlock()
	for _, program := range this.programs {
		if program.Name == name {
			return program
		}
	}

	return nil
}

func (this *dagrState) AddExecution(program *Program) *Execution {
	this.Lock()
	defer this.Unlock()
	id := uuid.New()
	execution := &Execution{id: id, program: program, subscribers: make(map[*websocket.Conn]bool)}
	this.executions[id] = execution
	return execution
}

func (this *dagrState) FindExecution(executionId string) *Execution {
	this.RLock()
	defer this.RUnlock()
	return this.executions[executionId]
}

func (this *dagrState) AllPrograms() []*Program {
	this.RLock()
	defer this.RUnlock()
	return this.programs
}

func MakeDagr(repo, workingDir string, delay time.Duration) (*dagrState, error) {
	s := &dagrState{executions: make(map[string]*Execution)}

	err := PullOrClone(repo, workingDir)

	if err != nil {
		return nil, err
	}

	getNewPrograms := func(sha string) (string, error) {
		newSha, err := MasterSha(repo)

		if err != nil {
			return sha, err
		}

		if newSha == sha {
			return sha, nil
		}

		err = Pull(workingDir)

		if err != nil {
			return sha, err
		}

		programs, err := readDir(workingDir)

		if err != nil {
			return sha, err
		}

		s.Lock()
		defer s.Unlock()
		s.programs = programs

		return newSha, nil
	}

	go func() {
		sha := ""

		for {
			newSha, err := getNewPrograms(sha)
			sha = newSha

			if err != nil {
				log.Println(err)
			}

			time.Sleep(delay)
		}
	}()

	return s, nil
}
