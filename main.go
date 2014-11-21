package main

import (
	"gopkg.in/alecthomas/kingpin.v1"
	"log"
)

var httpAddr = kingpin.Flag("http", "serve http on host:port").Short('a').Required().TCP()
var programsRepo = kingpin.Flag("repo", "repository containing programs").Short('r').Required().String()
var workingDir = kingpin.Flag("work", "working directory").Short('w').Required().String()

//var dbFile = kingpin.Flag("db", "sqlite database file").Short('d').Required().String()

func main() {
	kingpin.Parse()

	err := Update(*programsRepo, *workingDir)

	if err != nil {
		log.Fatal(err)
	}

	programs, err := ReadDir(*workingDir)

	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(Serve(httpAddr.String(), programs))
}
