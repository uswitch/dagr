package main

import (
	"gopkg.in/alecthomas/kingpin.v1"
	"log"
)

var httpAddr = kingpin.Flag("http", "serve http on host:port").Short('a').Required().TCP()
var programsRepo = kingpin.Flag("repo", "repository containing programs").Short('r').Required().String()
var workingDir = kingpin.Flag("work", "working directory").Short('w').Required().String()
var monitorInterval = kingpin.Flag("interval", "interval between checks for new programs").Short('i').Default("10s").Duration()

//var dbFile = kingpin.Flag("db", "sqlite database file").Short('d').Required().String()

func main() {
	kingpin.Parse()

	err := PullOrClone(*programsRepo, *workingDir)

	if err != nil {
		log.Fatal(err)
	}

	programs, err := MonitorPrograms(*programsRepo, *workingDir, *monitorInterval)

	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(Serve(httpAddr.String(), programs))
}
