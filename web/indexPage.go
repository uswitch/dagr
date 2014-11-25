package web

import (
	app "github.com/uswitch/dagr/app"
	"github.com/uswitch/dagr/program"
	"log"
	"net/http"
	"text/template"
)

type programStatus struct {
	Program           *program.Program
	LastExecution     *program.Execution
	LastExecutionTime string
	Running           bool
	Succeeded         bool
	Failed            bool
	Retryable         bool
}

type indexPageState struct {
	Succeeded       int
	Retryable       int
	Failed          int
	ProgramStatuses []*programStatus
}

func handleIndex(app app.App, indexTemplate *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		programs := app.Programs()
		programStatuses := []*programStatus{}

		var totalSucceeded, totalFailed, totalRetryable int

		for _, p := range programs {
			executions := p.Executions()
			var lastExecution *program.Execution
			var lastExecutionTime string
			if len(executions) > 0 {
				lastExecution = executions[len(executions)-1]
				lastExecutionTime = lastExecution.StartTime.Format("2 Jan 2006 15:04")
			}

			var running, succeeded, retryable, failed bool

			if lastExecution != nil {
				running = !lastExecution.Finished()

				if !running {
					succeeded = lastExecution.ExitStatus() == program.Success
					retryable = lastExecution.ExitStatus() == program.Retryable
					failed = lastExecution.ExitStatus() == program.Failed
				}
			}

			programStatuses = append(programStatuses,
				&programStatus{
					Program:           p,
					LastExecution:     lastExecution,
					LastExecutionTime: lastExecutionTime,
					Running:           running,
					Succeeded:         succeeded,
					Retryable:         retryable,
					Failed:            failed,
				})

			if succeeded {
				totalSucceeded++
			}

			if retryable {
				totalRetryable++
			}

			if failed {
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
