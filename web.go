package main

import (
	"bitbucket.org/tebeka/nrsc"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"regexp"
)

var TMPL = regexp.MustCompile(".tmpl$")

func handleIndex(programs []*Program) func(http.ResponseWriter, *http.Request) {
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

func handleExecution(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	programName := vars["program"]
	log.Println("executing program:", programName)
	http.Redirect(w, req, "/", 302)
}

func Serve(httpAddr string, programs []*Program) error {
	nrsc.Handle("/static/")
	nrsc.Mask(TMPL)

	r := mux.NewRouter()
	r.HandleFunc("/", handleIndex(programs))
	r.HandleFunc("/execute/{program}", handleExecution)
	http.Handle("/", r)

	server := &http.Server{
		Addr: httpAddr,
	}

	log.Println("dagr listening on", httpAddr)

	return server.ListenAndServe()
}
