package main

import (
	"bitbucket.org/tebeka/nrsc"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"regexp"
)

var TMPL = regexp.MustCompile(".tmpl$")

var index = template.Must(nrsc.LoadTemplates(nil, "index.html.tmpl"))
var info = template.Must(nrsc.LoadTemplates(nil, "info.html.tmpl"))

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
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		programName := vars["program"]
		program := dagr.FindProgram(programName)
		if program == nil {
			log.Println("no such program:", programName)
			http.NotFound(w, req)
		} else {
			if err := info.Execute(w, InfoState{program}); err != nil {
				http.NotFound(w, req)
			}
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
			program.Execute()
			http.Redirect(w, req, "/", 302)
		}
	}
}

func Serve(httpAddr string, dagr Dagr) error {
	nrsc.Handle("/static/")
	nrsc.Mask(TMPL)

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
