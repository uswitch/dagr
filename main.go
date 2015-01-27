package main

import (
	"github.com/uswitch/dagr/app"
	"github.com/uswitch/dagr/web"
	"gopkg.in/alecthomas/kingpin.v1"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
)

var httpAddr = kingpin.Flag("http", "serve http on host:port").Short('a').Required().TCP()
var programsRepo = kingpin.Flag("repo", "repository containing programs").Short('r').Required().String()
var workingDir = kingpin.Flag("work", "working directory").Short('w').Required().String()
var uiDir = kingpin.Flag("ui", "ui directory containing templates, js, css etc").Short('u').Required().String()
var monitorInterval = kingpin.Flag("interval", "interval between checks for new programs").Short('i').Default("10s").Duration()

// set during build
var Revision string

func trapSignal(app app.App, c <-chan os.Signal) {
	<-c
	log.Println("shutting down")
	app.Shutdown()
	log.Println("finished shutdown, exiting")
	os.Exit(0)
}

func main() {
	kingpin.Parse()

	log.Println("dagr", Revision)
	log.Println("starting application")
	app, err := app.New(*programsRepo, *workingDir)

	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func() {
		trapSignal(app, c)
	}()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(*uiDir, "static")))))
	http.Handle("/", web.DagrHandler(app, filepath.Join(*uiDir, "templates")))

	server := &http.Server{
		Addr: (*httpAddr).String(),
	}

	log.Println("dagr listening on", *httpAddr)

	app.Run(*monitorInterval)

	log.Fatal(server.ListenAndServe())
}
