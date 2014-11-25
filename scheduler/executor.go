package scheduler

import (
	"github.com/uswitch/dagr/program"
	"log"
)

type Executor struct {
	executionRequests  chan *executionRequest
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
	return &Executor{
		executionRequests:  make(chan *executionRequest),
		executions:         make(chan *program.Execution),
		recordedExecutions: make(map[string]*program.Execution),
	}
}

func (e *Executor) FindExecution(executionId string) *program.Execution {
	return e.recordedExecutions[executionId]
}

func (e *Executor) doExecute(er *executionRequest) (*program.Execution, error) {
	log.Println("executing", er.program.Name)

	execution, err := er.program.Execute()

	log.Println(er.program.Name, "executed")

	if err != nil {
		log.Println("execution error", err)
		er.result <- &executionResult{nil, err}
		// record error as well?
		return nil, err
	}

	log.Println("execution ok -- sending result to channel")
	er.result <- &executionResult{execution, nil}

	log.Println("result sent")
	e.recordedExecutions[execution.Id] = execution
	return execution, nil
}

func (e *Executor) Execute(p *program.Program) (*program.Execution, error) {
	ch := make(chan *executionResult)
	log.Println("sending execution request for program", p.Name)
	e.executionRequests <- &executionRequest{p, ch}
	log.Println("request sent -- awaiting result")
	result := <-ch
	log.Println("got result")
	return result.execution, result.err
}

func (e *Executor) RunExecutorLoop() {
	log.Println("running executor loop")
	for er := range e.executionRequests {
		log.Println("got an execution request")
		execution, err := e.doExecute(er)
		if err == nil {
			e.executions <- execution
			e.recordedExecutions[execution.Id] = execution
		} else {
			log.Println(err)
		}
	}
}
