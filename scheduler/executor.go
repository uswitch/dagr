package scheduler

import (
	"github.com/uswitch/dagr/program"
	"log"
)

type Executor struct {
	programs           chan *executionRequest
	executions         chan *program.Execution
	recordedExecutions map[string]*program.Execution
}

type executionResult struct {
	execution *program.Execution
	err       error
}

type executionRequest struct {
	program *program.Program
	result  chan *executionResult
}

func NewExecutor() *Executor {
	return &Executor{}
}

func (e *Executor) FindExecution(executionId string) *program.Execution {
	return e.recordedExecutions[executionId]
}

func (e *Executor) doExecute(er *executionRequest) (*program.Execution, error) {
	execution, err := er.program.Execute()

	if err != nil {
		er.result <- &executionResult{nil, err}
		// record error as well?
		return nil, err
	}

	er.result <- &executionResult{execution, nil}
	e.recordedExecutions[execution.Id] = execution
	return execution, nil
}

func (e *Executor) Execute(p *program.Program) (*program.Execution, error) {
	ch := make(chan *executionResult)
	e.programs <- &executionRequest{p, ch}
	result := <-ch
	return result.execution, result.err
}

func (e *Executor) RunExecutorLoop() {
	for er := range e.programs {
		execution, err := e.doExecute(er)
		if err == nil {
			e.executions <- execution
			e.recordedExecutions[execution.Id] = execution
		} else {
			log.Println(err)
		}
	}
}
