package main

import (
	"bitbucket.org/tebeka/nrsc"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"regexp"
)

var TMPL = regexp.MustCompile(".tmpl$")

type Status struct {
	Succeeded int
	Retryable int
	Failed    int
	Programs  []*Program
}

func handleIndex(dagr Dagr) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		t, err := nrsc.LoadTemplates(nil, "index.html.tmpl")
		if err != nil {
			http.NotFound(w, req)
		}
		if err = t.Execute(w, Status{77, 13, 12, dagr.AllPrograms()}); err != nil {
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
			log.Println("executing program:", program)
			program.Execute()
			http.Redirect(w, req, "/", 302)
		}
	}
}

func Serve(httpAddr string, dagr Dagr) error {
	nrsc.Handle("/static/")
	nrsc.Mask(TMPL)

	r := mux.NewRouter()
	r.HandleFunc("/", handleIndex(dagr))
	r.HandleFunc("/execute/{program}", handleExecution(dagr))
	http.Handle("/", r)

	server := &http.Server{
		Addr: httpAddr,
	}

	log.Println("dagr listening on", httpAddr)

	return server.ListenAndServe()
}
