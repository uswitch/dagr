package app

import (
	"github.com/uswitch/dagr/program"
	"github.com/uswitch/dagr/scheduler"
	"time"
)

type App interface {
	Execute(*program.Program) (*program.Execution, error)
	Programs() []*program.Program
	FindProgram(string) *program.Program
	FindExecution(string) *program.Execution
}

type appState struct {
	executor   *scheduler.Executor
	repository *program.Repository
}

func (a *appState) Execute(p *program.Program) (*program.Execution, error) {
	execution, err := a.executor.Execute(p)
	return execution, err
}

func (a *appState) FindExecution(executionId string) *program.Execution {
	return a.executor.FindExecution(executionId)
}

func (a *appState) FindProgram(name string) *program.Program {
	return a.repository.FindProgram(name)
}

func (a *appState) Programs() []*program.Program {
	return a.repository.Programs()
}

func New(repo, workingDir string, delay time.Duration) (*appState, error) {
	executor := scheduler.NewExecutor()
	repository, err := program.NewRepository(repo, workingDir)

	if err != nil {
		return nil, err
	}

	return &appState{executor: executor, repository: repository}, nil
}

func (a *appState) Run() {
	go a.executor.RunExecutorLoop()
	go a.repository.RunRefreshLoop(time.Tick(60 * time.Second))
	go scheduler.RunScheduleLoop(a.repository, a.executor, time.Tick(1*time.Second))
}
