package web

import (
	"github.com/gorilla/mux"
	"github.com/uswitch/dagr/app"
	"github.com/uswitch/dagr/program"
	"log"
	"net/http"
	"text/template"
)

type programPageState struct {
	Program *program.Program
}

func handleProgramInfo(app app.App, infoTemplate *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		programName := vars["program"]
		program := app.FindProgram(programName)
		if program == nil {
			log.Println("no such program:", programName)
			http.NotFound(w, req)
		} else if err := infoTemplate.Execute(w, programPageState{program}); err != nil {
			log.Println("error when executing program info template:", err)
			http.Error(w, err.Error(), 500)
		}
	}
}
