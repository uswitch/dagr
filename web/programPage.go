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

type programPageState struct {
	Program                     *program.Program
	ExecutionStatuses           []*executionStatus
	ProgramExecutionsSocketPath string
}

func handleProgramInfo(app app.App, infoTemplate *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		programName := vars["program"]
		program := app.FindProgram(programName)
		if program == nil {
			log.Println("no such program:", programName)
			http.NotFound(w, req)
		} else {
			executionStatuses := []*executionStatus{}

			for _, e := range program.Executions() {
				executionStatuses = append(executionStatuses, newExecutionStatus(e))
			}

			programExecutionsSocketPath := fmt.Sprintf("/program/%s/executions", program.Name)

			if err := infoTemplate.Execute(w, programPageState{program, executionStatuses, programExecutionsSocketPath}); err != nil {
				log.Println("error when executing program info template:", err)
				http.Error(w, err.Error(), 500)
			}
		}
	}
}
