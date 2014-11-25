package web

import (
	"github.com/gorilla/mux"
	dagr "github.com/uswitch/dagr/dagrpkg"
	"log"
	"net/http"
)

func handleProgramExecute(dagr dagr.Dagr) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		programName := vars["program"]
		program := dagr.FindProgram(programName)
		if program == nil {
			log.Println("no such program:", programName)
			http.NotFound(w, req)
		} else {
			execution, err := dagr.Execute(program)

			if err != nil {
				log.Println("error on execution:", err)
				http.Error(w, err.Error(), 500)
				return
			}

			http.Redirect(w, req, "/executions/"+execution.Id, 302)
		}
	}
}
