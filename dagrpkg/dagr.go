package dagrpkg

import (
	"github.com/uswitch/dagr/program"
	"github.com/uswitch/dagr/scheduler"
	"time"
)

type Dagr interface {
	Execute(*program.Program) (*program.Execution, error)
	Programs() []*program.Program
	FindProgram(string) *program.Program
	FindExecution(string) *program.Execution
}

type dagrState struct {
	executor   *scheduler.Executor
	repository *program.Repository
}

func (this *dagrState) Execute(p *program.Program) (*program.Execution, error) {
	execution, err := this.executor.Execute(p)
	return execution, err
}

func (this *dagrState) FindExecution(executionId string) *program.Execution {
	return this.executor.FindExecution(executionId)
}

func (this *dagrState) FindProgram(name string) *program.Program {
	return this.repository.FindProgram(name)
}

func (this *dagrState) Programs() []*program.Program {
	return this.repository.Programs()
}

func New(repo, workingDir string, delay time.Duration) (*dagrState, error) {
	executor := scheduler.NewExecutor()
	repository, err := program.NewRepository(repo, workingDir)

	if err != nil {
		return nil, err
	}

	return &dagrState{executor: executor, repository: repository}, nil
}

func (d *dagrState) Run() {
	go d.repository.RunRefreshLoop(time.Tick(60 * time.Second))
	go scheduler.RunScheduleLoop(d.repository, d.executor, time.Tick(1*time.Second))
}
