package main

import (
	"github.com/GeertJohan/go.rice"
	"gopkg.in/alecthomas/kingpin.v1"
	"log"
	"net/http"
)

var httpAddr = kingpin.Flag("http", "serve http on host:port").Short('a').Required().TCP()
var programsRepo = kingpin.Flag("repo", "repository containing programs").Short('r').Required().String()
var workingDir = kingpin.Flag("work", "working directory").Short('w').Required().String()
var monitorInterval = kingpin.Flag("interval", "interval between checks for new programs").Short('i').Default("10s").Duration()

//var dbFile = kingpin.Flag("db", "sqlite database file").Short('d').Required().String()

func main() {
	kingpin.Parse()

	dagr, err := MakeDagr(*programsRepo, *workingDir, *monitorInterval)

	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(rice.MustFindBox("resources/static").HTTPBox())))
	http.Handle("/", DagrHandler(dagr))

	server := &http.Server{
		Addr: httpAddr.String(),
	}

	log.Println("dagr listening on", *httpAddr)

	log.Fatal(server.ListenAndServe())
}
