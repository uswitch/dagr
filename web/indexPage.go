package web

import (
	"github.com/uswitch/dagr/app"
	"github.com/uswitch/dagr/program"
	"log"
	"net/http"
	"text/template"
)

type programStatus struct {
	Program *program.Program
	*executionStatus
}

type indexPageState struct {
	Succeeded       int
	Retryable       int
	Failed          int
	ProgramStatuses []*programStatus
}

func newExecutionStatus(execution *program.Execution) *executionStatus {
	var executionTime string
	var running, succeeded, retryable, failed bool

	if execution != nil {
		executionTime = execution.StartTime.Format("2 Jan 2006 15:04")
		running = !execution.Finished()

		if !running {
			succeeded = execution.ExitStatus() == program.Success
			retryable = execution.ExitStatus() == program.Retryable
			failed = execution.ExitStatus() == program.Failed
		}
	}

	return &executionStatus{
		Execution:     execution,
		ExecutionTime: executionTime,
		Running:       running,
		Succeeded:     succeeded,
		Retryable:     retryable,
		Failed:        failed,
	}
}

func handleIndex(app app.App, indexTemplate *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		programs := app.Programs()
		programStatuses := []*programStatus{}

		var totalSucceeded, totalFailed, totalRetryable int

		for _, p := range programs {
			executions := p.Executions()
			var lastExecution *program.Execution
			if len(executions) > 0 {
				lastExecution = executions[len(executions)-1]
			}
			executionStatus := newExecutionStatus(lastExecution)

			programStatuses = append(programStatuses,
				&programStatus{
					Program:         p,
					executionStatus: executionStatus,
				})

			if executionStatus.Succeeded {
				totalSucceeded++
			}

			if executionStatus.Retryable {
				totalRetryable++
			}

			if executionStatus.Failed {
				totalFailed++
			}
		}

		err := indexTemplate.Execute(w, indexPageState{totalSucceeded, totalRetryable, totalFailed, programStatuses})

		if err != nil {
			log.Println("error when executing index template:", err)
			http.Error(w, err.Error(), 500)
		}
	}
}
