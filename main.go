package main

import (
	"gopkg.in/alecthomas/kingpin.v1"
	"log"
	"net/http"
)

var httpAddr = kingpin.Flag("http", "serve http on host:port").Default(":7730").Short('a').Required().TCP()
var tasksRepo = kingpin.Flag("repo", "repository containing tasks").Short('r').Required().String()
var workingDir = kingpin.Flag("work", "working directory").Short('w').Required().String()
var dbFile = kingpin.Flag("db", "sqlite database file").Short('d').Required().String()

func main() {
	kingpin.Parse()

	server := &http.Server{
		Addr: httpAddr.String(),
	}

	log.Fatal(server.ListenAndServe())
}
