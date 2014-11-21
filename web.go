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

func handleIndex(ch chan []*Program) func(http.ResponseWriter, *http.Request) {
	programs := []*Program{}

	go func() {
		for {
			select {
			case newPrograms := <-ch:
				programs = newPrograms //???
			}
		}
	}()

	return func(w http.ResponseWriter, req *http.Request) {
		t, err := nrsc.LoadTemplates(nil, "index.html.tmpl")
		if err != nil {
			http.NotFound(w, req)
		}
		if err = t.Execute(w, Status{77, 13, 12, programs}); err != nil {
			http.NotFound(w, req)
		}
	}
}

func handleExecution(ch chan []*Program) func(http.ResponseWriter, *http.Request) {
	programs := []*Program{}

	go func() {
		for {
			select {
			case newPrograms := <-ch:
				programs = newPrograms //???
			}
		}
	}()

	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		programName := vars["program"]
		program := FindProgram(programName, programs)
		if program == nil {
			log.Println("no such program:", programName)
			http.NotFound(w, req)
		} else {
			log.Println("executing program:", program)
			http.Redirect(w, req, "/", 302)
		}
	}
}

func Serve(httpAddr string, programs chan []*Program) error {
	nrsc.Handle("/static/")
	nrsc.Mask(TMPL)

	indexCh := make(chan []*Program)
	executionCh := make(chan []*Program)

	go func() {
		for {
			select {
			case newPrograms := <-programs:
				indexCh <- newPrograms     //???
				executionCh <- newPrograms //???
			}
		}
	}()

	r := mux.NewRouter()
	r.HandleFunc("/", handleIndex(indexCh))
	r.HandleFunc("/execute/{program}", handleExecution(executionCh))
	http.Handle("/", r)

	server := &http.Server{
		Addr: httpAddr,
	}

	log.Println("dagr listening on", httpAddr)

	return server.ListenAndServe()
}
