package main

import (
	"log"
	"sync"
	"time"
)

type Dagr interface {
	AllPrograms() []*Program
	FindProgram(string) *Program
}

type dagrState struct {
	programs []*Program
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

func (this *dagrState) AllPrograms() []*Program {
	this.RLock()
	defer this.RUnlock()
	return this.programs
}

func MakeDagr(repo, workingDir string, delay time.Duration) (*dagrState, error) {
	s := &dagrState{}

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
