package scheduler

import (
	"github.com/uswitch/dagr/program"
	"log"
	"time"
)

func RunScheduleLoop(repository *program.Repository, executor *Executor, ticks <-chan time.Time, shutdown <-chan bool) {
	for {
		select {
		case now := <-ticks:
			for _, p := range selectExecutablePrograms(repository.Programs(), now) {
				program.ProgramLog(p, "scheduling execution")
				//execution, err :=
				executor.Execute(p)
			}
		case _ = <-shutdown:
			log.Println("shutting down scheduler")
			return
		}
	}
}

func selectExecutablePrograms(programs []*program.Program, instant time.Time) []*program.Program {
	readyPrograms := []*program.Program{}

	for _, p := range programs {
		if isReady(p, instant) {
			readyPrograms = append(readyPrograms, p)
		}
	}

	return readyPrograms
}

func isReady(p *program.Program, instant time.Time) bool {
	executions := p.Executions()

	if len(executions) == 0 {
		// never run, therefore ready
		return true
	}

	lastExecution := executions[len(executions)-1]

	if !lastExecution.Finished() {
		// still running, therefore not ready
		return false
	}

	// ready if ended before 'instant' and next scheduled execution time before 'instant'

	lastExecutionEndTime := lastExecution.StartTime.Add(lastExecution.Duration())
	nextExecutionStartTime := p.Schedule.Next(lastExecutionEndTime)
	return lastExecutionEndTime.Before(instant) && nextExecutionStartTime.Before(instant)
}
