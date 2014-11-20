package main

import (
	"bitbucket.org/tebeka/nrsc"
	"gopkg.in/alecthomas/kingpin.v1"
	"log"
	"net/http"
	"regexp"
)

var TMPL = regexp.MustCompile(".tmpl$")

var httpAddr = kingpin.Flag("http", "serve http on host:port").Short('a').Required().TCP()

//var tasksRepo = kingpin.Flag("repo", "repository containing tasks").Short('r').Required().String()
//var workingDir = kingpin.Flag("work", "working directory").Short('w').Required().String()
//var dbFile = kingpin.Flag("db", "sqlite database file").Short('d').Required().String()

type Status struct {
	Succeeded int
	Retryable int
	Failed    int
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	t, err := nrsc.LoadTemplates(nil, "index.html.tmpl")
	if err != nil {
		http.NotFound(w, req)
	}
	if err = t.Execute(w, Status{77, 13, 12}); err != nil {
		http.NotFound(w, req)
	}
}

func main() {
	kingpin.Parse()

	nrsc.Handle("/static/")
	nrsc.Mask(TMPL)
	http.HandleFunc("/", indexHandler)

	server := &http.Server{
		Addr: httpAddr.String(),
	}

	log.Fatal(server.ListenAndServe())
}
