package web

import (
	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	"github.com/uswitch/dagr/app"
	"net/http"
	"text/template"
)

func DagrHandler(app app.App, templates *rice.Box) http.Handler {
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

func loadTemplate(templates *rice.Box, path string) *template.Template {
	templateString, err := templates.String(path)
	if err != nil {
		panic(err)
	}
	return template.Must(template.New(path).Parse(templateString))
}
