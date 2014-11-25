package web

import (
	"github.com/gorilla/mux"
	dagr "github.com/uswitch/dagr/dagrpkg"
	"github.com/uswitch/dagr/program"
	"log"
	"net/http"
	"text/template"
)

type executionPageState struct {
	Execution *program.Execution
}

func handleExecutionInfo(dagr dagr.Dagr, showTemplate *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		executionId := vars["executionId"]
		execution := dagr.FindExecution(executionId)
		if execution == nil {
			log.Println("no such execution:", executionId)
			http.NotFound(w, req)
		} else if err := showTemplate.Execute(w, executionPageState{execution}); err != nil {
			log.Println("error when executing execution info template:", err)
			http.Error(w, err.Error(), 500)
		}
	}
}
