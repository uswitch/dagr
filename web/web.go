package web

import (
	"github.com/gorilla/mux"
	"github.com/uswitch/dagr/app"
	"net/http"
	"path/filepath"
	"text/template"
)

func DagrHandler(app app.App, templates string) http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", handleIndex(app, loadTemplate(templates, "index.html.tmpl"))).Methods("GET")
	r.HandleFunc("/program/{program}", handleProgramInfo(app, loadTemplate(templates, "program.html.tmpl"))).Methods("GET")
	r.HandleFunc("/program/{program}/execute", handleProgramExecute(app)).Methods("POST")
	r.HandleFunc("/program/{program}/executions", programExecutions(app))
	r.HandleFunc("/executions/{executionId}",
		handleExecutionInfo(app, loadTemplate(templates, "execution.html.tmpl"))).Methods("GET")
	r.HandleFunc("/executions/{executionId}/messages/{countSoFar:[0-9]+}", handleExecutionMessages(app))
	return r
}

func loadTemplate(path, filename string) *template.Template {
	return template.Must(template.ParseFiles(filepath.Join(path, filename)))
}
