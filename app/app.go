package app

import (
	"github.com/uswitch/dagr/program"
	"github.com/uswitch/dagr/scheduler"
	"log"
	"sync"
	"time"
)

type App interface {
	Execute(*program.Program) (*program.Execution, error)
	Programs() []*program.Program
	FindProgram(string) *program.Program
	FindExecution(string) *program.Execution
	Run(time.Duration)
	Shutdown()
}

type appState struct {
	executor            *scheduler.Executor
	repository          *program.Repository
	schedulerShutdownCh chan bool
	waitGroup           *sync.WaitGroup
}

func (a *appState) Execute(p *program.Program) (*program.Execution, error) {
	return a.executor.Execute(p)
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

func (a *appState) Shutdown() {
	log.Println("shutting down application")
	a.schedulerShutdownCh <- true
	a.executor.Shutdown()
	a.waitGroup.Wait()
	log.Println("finished application shutdown")
}

func New(repo, workingDir string) (*appState, error) {
	log.Println("starting executor")
	executor := scheduler.NewExecutor()
	log.Println("initialising programs repository")
	repository, err := program.NewRepository(repo, workingDir)

	if err != nil {
		return nil, err
	}

	return &appState{executor: executor, repository: repository, waitGroup: &sync.WaitGroup{}, schedulerShutdownCh: make(chan bool)}, nil
}

func (a *appState) Run(repositoryCheckInterval time.Duration) {
	go a.executor.RunExecutorLoop()
	go a.repository.RunRefreshLoop(time.Tick(repositoryCheckInterval))	

	a.waitGroup.Add(1)
	go func() {
		scheduler.RunScheduleLoop(a.repository, a.executor, time.Tick(1*time.Second), a.schedulerShutdownCh)
		a.waitGroup.Done()
	}()
}
