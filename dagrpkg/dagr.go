package dagrpkg

import (
	"github.com/uswitch/dagr/git"
	"github.com/uswitch/dagr/program"
	"log"
	"sync"
	"time"
)

type Dagr interface {
	Programs() []*program.Program
	Execute(*program.Program) (*program.Execution, error)
	FindProgram(string) *program.Program
	FindExecution(string) *program.Execution
}

type dagrState struct {
	programs   []*program.Program
	executions map[string]*program.Execution
	sync.RWMutex
}

func (this *dagrState) Execute(program *program.Program) (*program.Execution, error) {
	this.Lock()
	defer this.Unlock()
	execution, err := program.Execute()

	if err != nil {
		return nil, err
	}

	this.executions[execution.Id] = execution
	return execution, nil
}

func (this *dagrState) FindProgram(name string) *program.Program {
	this.RLock()
	defer this.RUnlock()
	for _, program := range this.programs {
		if program.Name == name {
			return program
		}
	}

	return nil
}

func (this *dagrState) FindExecution(executionId string) *program.Execution {
	this.RLock()
	defer this.RUnlock()
	return this.executions[executionId]
}

func (this *dagrState) Programs() []*program.Program {
	this.RLock()
	defer this.RUnlock()
	return this.programs
}

func New(repo, workingDir string, delay time.Duration) (*dagrState, error) {
	s := &dagrState{executions: make(map[string]*program.Execution)}

	err := git.PullOrClone(repo, workingDir)

	if err != nil {
		return nil, err
	}

	getNewPrograms := func(sha string) (string, error) {
		newSha, err := git.MasterSha(repo)

		if err != nil {
			return sha, err
		}

		if newSha == sha {
			return sha, nil
		}

		err = git.Pull(workingDir)

		if err != nil {
			return sha, err
		}

		programs, err := program.ReadDir(workingDir)

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
