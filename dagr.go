package main

import (
	"log"
	"sync"
	"time"
)

type Dagr interface {
	AllPrograms() []*Program
	AddExecution(string, *Execution)
	FindProgram(string) *Program
	FindExecution(string) *Execution
}

type dagrState struct {
	programs   []*Program
	executions map[string]*Execution
	sync.RWMutex
}

func newDagrState() *dagrState {
	s := &dagrState{}
	s.executions = make(map[string]*Execution)
	return s
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

func (this *dagrState) AddExecution(executionId string, execution *Execution) {
	this.Lock()
	defer this.Unlock()
	this.executions[executionId] = execution
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
	s := newDagrState()

	err := PullOrClone(repo, workingDir)

	if err != nil {
		return nil, err
	}

	sha := ""

	go func() {
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
