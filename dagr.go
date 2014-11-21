package main

import (
	"log"
	"time"
)

type Dagr interface {
	AllPrograms() []*Program
	FindProgram(string) *Program
}

type dagrState struct {
	programs []*Program
}

func (this *dagrState) FindProgram(name string) *Program {
	for _, program := range this.programs {
		if program.Name == name {
			return program
		}
	}

	return nil
}

func (this *dagrState) AllPrograms() []*Program {
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

				s.programs = newPrograms
				sha = newSha
			}
		}
	}()

	return s, nil
}
