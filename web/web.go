package web

import (
	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	dagr "github.com/uswitch/dagr/dagrpkg"
	"net/http"
	"text/template"
)

func DagrHandler(dagr dagr.Dagr, templates *rice.Box) http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", handleIndex(dagr, loadTemplate(templates, "index.html.tmpl"))).Methods("GET")
	r.HandleFunc("/program/{program}", handleProgramInfo(dagr, loadTemplate(templates, "program.html.tmpl"))).Methods("GET")
	r.HandleFunc("/program/{program}/execute", handleProgramExecute(dagr)).Methods("POST")
	r.HandleFunc("/executions/{executionId}",
		handleExecutionInfo(dagr, loadTemplate(templates, "execution.html.tmpl"))).Methods("GET")
	r.HandleFunc("/executions/{executionId}/messages/{countSoFar:[0-9]+}", handleExecutionMessages(dagr))
	return r
}

func loadTemplate(templates *rice.Box, path string) *template.Template {
	templateString, err := templates.String(path)
	if err != nil {
		panic(err)
	}
	return template.Must(template.New(path).Parse(templateString))
}
