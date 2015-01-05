package web

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/uswitch/dagr/app"
	"github.com/uswitch/dagr/program"
	"log"
	"net/http"
	"strconv"
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
			
			limit := 10
			queryLimit := req.URL.Query().Get("limit")
			if queryLimit != "" {
				n, err := strconv.ParseInt(queryLimit, 10, 0)
				if err == nil {
					limit = int(n)
				}
			}
			
			lo := 0
			if len(program.Executions()) > limit {
				lo = len(program.Executions())-limit
			}
			
			for _, e := range program.Executions()[lo:] {
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
