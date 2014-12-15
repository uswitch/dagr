package web

import (
	"fmt"
	"github.com/uswitch/dagr/app"
	"github.com/uswitch/dagr/program"
	"log"
	"net/http"
	"text/template"
)

type programStatus struct {
	Program *program.Program
	*executionStatus
	ProgramExecutionsSocketPath string
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
			succeeded = execution.Status() == program.SuccessStatus
			retryable = execution.Status() == program.RetryableStatus
			failed = execution.Status() == program.FailedStatus
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

			programExecutionsSocketPath := fmt.Sprintf("/program/%s/executions", p.Name)

			programStatuses = append(programStatuses,
				&programStatus{
					Program:                     p,
					ProgramExecutionsSocketPath: programExecutionsSocketPath,
					executionStatus:             executionStatus,
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
