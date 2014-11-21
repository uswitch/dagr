package main

import (
	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"regexp"
	"text/template"
)

var TMPL = regexp.MustCompile(".tmpl$")

var resourceBox = rice.MustFindBox("resources")

type IndexState struct {
	Succeeded int
	Retryable int
	Failed    int
	Programs  []*Program
}

type InfoState struct {
	Program *Program
}

func handleIndex(dagr Dagr) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if err := index.Execute(w, IndexState{77, 13, 12, dagr.AllPrograms()}); err != nil {
			http.NotFound(w, req)
		}
	}
}

func handleInfo(dagr Dagr) func(http.ResponseWriter, *http.Request) {
	var infoTemplate, err = loadTemplate("info.html.tmpl")

	if err != nil {
		log.Fatal(err)
	}

	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		programName := vars["program"]
		program := dagr.FindProgram(programName)
		if program == nil {
			log.Println("no such program:", programName)
			http.NotFound(w, req)
		} else if err := infoTemplate.Execute(w, InfoState{program}); err != nil {
			http.NotFound(w, req)
		}
	}
}

func handleExecution(dagr Dagr) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		programName := vars["program"]
		program := dagr.FindProgram(programName)
		if program == nil {
			log.Println("no such program:", programName)
			http.NotFound(w, req)
		} else {
			exec := NewExecution(program)
			exec.Execute()
			http.Redirect(w, req, "/", 302)
		}
	}
}

func loadTemplate(path string) (*template.Template, error) {
	templateString, err := resourceBox.String(path)
	if err != nil {
		return nil, err
	}
	var t = template.New(path)
	return t.Parse(templateString)
}

func Serve(httpAddr string, dagr Dagr) error {

	var executionTemplate = template.New("execution.html.tmpl")
	var indexTemplate = template.New("index.html.tmpl")

	r := mux.NewRouter()
	r.HandleFunc("/", handleIndex(dagr)).Methods("GET")
	r.HandleFunc("/program/{program}", handleInfo(dagr)).Methods("GET")
	r.HandleFunc("/program/{program}/execute", handleExecution(dagr)).Methods("POST")
	http.Handle("/", r)

	server := &http.Server{
		Addr: httpAddr,
	}

	log.Println("dagr listening on", httpAddr)

	return server.ListenAndServe()
}
