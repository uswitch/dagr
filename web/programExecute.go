package web

import (
	"github.com/gorilla/mux"
	"github.com/uswitch/dagr/app"
	"log"
	"net/http"
)

func handleProgramExecute(app app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		programName := vars["program"]
		program := app.FindProgram(programName)
		if program == nil {
			log.Println("no such program:", programName)
			http.NotFound(w, req)
		} else {
			execution, err := app.Execute(program)

			if err != nil {
				log.Println("error on execution:", err)
				http.Error(w, err.Error(), 500)
				return
			}

			http.Redirect(w, req, "/executions/"+execution.Id, 302)
		}
	}
}
