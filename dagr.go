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

	go func() {
		sha := ""

		for {
			defer func() {
				time.Sleep(delay)
			}()

			newSha, err := MasterSha(repo)

			if err != nil {
				log.Print(err)
				continue
			}

			if newSha != sha {
				err := Pull(workingDir)

				if err != nil {
					log.Print(err)
					continue
				}

				newPrograms, err := readDir(workingDir)

				if err != nil {
					log.Print(err)
					continue
				}

				s.Lock()
				s.programs = newPrograms
				s.Unlock()
				sha = newSha
			}
		}
	}()

	return s, nil
}
