package scheduler

import (
	"github.com/uswitch/dagr/program"
	"log"
	"time"
)

type Executor struct {
	executionRequests  chan *executionRequest
	recordedExecutions map[string]*program.Execution
	executionSlotAvailable chan bool
}

type executionResponse struct {
	execution *program.Execution
	exitCh    chan program.ExitCode
	err       error
}

type executionRequest struct {
	program    *program.Program
	responseCh chan *executionResponse
}

func NewExecutor(concurrentExecutions int) *Executor {
	slotCh := make(chan bool)
	go func() {
		for i := 0; i < concurrentExecutions; i++ {
			slotCh <- true
		}
	}()
	
	return &Executor{
		executionRequests:      make(chan *executionRequest, 100),
		recordedExecutions:     make(map[string]*program.Execution),
		executionSlotAvailable: slotCh,
	}
}

func (e *Executor) Shutdown() {	
	log.Println("stopping running executions")
	for _, execution := range e.recordedExecutions {
		if execution.IsRunning() {
			program.ExecutionLog(execution, "starting graceful shutdown")
			execution.Shutdown()
		}
	}
	log.Println("finished shutdown of all running executions")
}

func (e *Executor) FindExecution(executionId string) *program.Execution {
	return e.recordedExecutions[executionId]
}

func (e *Executor) doExecute(er *executionRequest) (*program.Execution, error) {
	program.ProgramLog(er.program, "executing")

	exitCh := make(chan program.ExitCode)
	execution, err := er.program.Execute(e.executionSlotAvailable, exitCh)

	if err != nil {
		if execution != nil {
			program.ExecutionLog(execution, "error", err)
		} else {
			program.ProgramLog(er.program, "error", err)
		}
		er.responseCh <- &executionResponse{nil, nil, err}
		// record error as well?
		return nil, err
	} else {
		program.ExecutionLog(execution, "started execution")
	}

	er.responseCh <- &executionResponse{execution, exitCh, nil}

	e.recordedExecutions[execution.Id] = execution
	return execution, nil
}

func (e *Executor) monitorExecution(pe *program.Execution, ch chan program.ExitCode) {
	program.ExecutionLog(pe, "monitoring execution")
	
	exitCode := <-ch
	e.slotAvailable()
	
	program.ExecutionLog(pe, "execution completed", exitCode)
	if exitCode == program.RetryableCode {
		program.ExecutionLog(pe, "scheduling for retry in 1m")
		time.Sleep(1 * time.Minute) // FIXME: make configurable
		_, err := e.Execute(pe.Program)
		if err != nil {
			log.Println(err)
		}
	}
}

func (e *Executor) Execute(p *program.Program) (*program.Execution, error) {
	ch := make(chan *executionResponse)

	e.executionRequests <- &executionRequest{p, ch}
	response := <-ch

	if response.err == nil {
		go e.monitorExecution(response.execution, response.exitCh)
	}

	return response.execution, response.err
}

func (e *Executor) slotAvailable() {
	go func() { e.executionSlotAvailable <- true }()
}

func (e *Executor) RunExecutorLoop() {
	for er := range e.executionRequests {
		program.ProgramLog(er.program, "got an execution request")
		
		execution, err := e.doExecute(er)
		if err == nil {
			e.recordedExecutions[execution.Id] = execution
		} else {
			program.ProgramLog(er.program, "error", err)
		}
	}
}
