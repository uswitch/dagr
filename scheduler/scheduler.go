package scheduler

import (
	"github.com/uswitch/dagr/program"
	"time"
)

func RunScheduleLoop(repository *program.Repository, executor *Executor, ticks <-chan time.Time) {
	for now := range ticks {
		for _, p := range selectExecutablePrograms(repository.Programs(), now) {
			program.ProgramLog(p, "scheduling execution")
			//execution, err :=
			executor.Execute(p)
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

	lastExecutionEndTime := lastExecution.StartTime.Add(lastExecution.Duration())

	// ready if ended before 'instant' and not on the same day as 'instant'
	return lastExecutionEndTime.Before(instant) && !sameDay(lastExecutionEndTime, instant)
}

func sameDay(time1, time2 time.Time) bool {
	y1, m1, d1 := time1.Date()
	y2, m2, d2 := time2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
