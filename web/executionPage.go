package web

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/uswitch/dagr/app"
	"github.com/uswitch/dagr/program"
	"log"
	"net/http"
	"text/template"
)

type executionStatus struct {
	Execution           *program.Execution
	ExecutionTime       string
	ExecutionLastOutput string
	Running             bool
	Succeeded           bool
	Failed              bool
	Retryable           bool
	Waiting             bool
}

type executionPageState struct {
	*executionStatus
	ProgramExecutionsSocketPath string
}

func handleExecutionInfo(app app.App, showTemplate *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		executionId := vars["executionId"]
		execution := app.FindExecution(executionId)
		if execution == nil {
			log.Println("no such execution:", executionId)
			http.NotFound(w, req)
		} else {
			programExecutionsSocketPath := fmt.Sprintf("/program/%s/executions", execution.Program.Name)

			if err := showTemplate.Execute(w, executionPageState{newExecutionStatus(execution), programExecutionsSocketPath}); err != nil {
				log.Println("error when executing execution info template:", err)
				http.Error(w, err.Error(), 500)
			}
		}
	}
}
