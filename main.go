package main

import (
	"bitbucket.org/tebeka/nrsc"
	"gopkg.in/alecthomas/kingpin.v1"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

var TMPL = regexp.MustCompile(".tmpl$")

var httpAddr = kingpin.Flag("http", "serve http on host:port").Short('a').Required().TCP()
var workingDir = kingpin.Flag("work", "working directory").Short('w').Required().String()

//var programsRepo = kingpin.Flag("repo", "repository containing programs").Short('r').Required().String()
//var dbFile = kingpin.Flag("db", "sqlite database file").Short('d').Required().String()

type Program struct {
	Name string `json:"name"`
}

type Status struct {
	Succeeded int
	Retryable int
	Failed    int
	Programs  []*Program
}

// does the given directory contain a 'main' file?
func isDagrProgram(path string) bool {
	_, err := os.Stat(filepath.Join(path, "main"))
	return err == nil
}

func readPrograms(dir string) ([]*Program, error) {
	programs := []*Program{}

	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return programs, err
	}

	for _, info := range infos {
		if err == nil && info.IsDir() && isDagrProgram(filepath.Join(dir, info.Name())) {
			programs = append(programs, &Program{info.Name()})
		}
	}
	return programs, nil
}

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

func main() {
	kingpin.Parse()

	programs, err := readPrograms(*workingDir)

	nrsc.Handle("/static/")
	nrsc.Mask(TMPL)
	http.HandleFunc("/", handleIndex(programs))

	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr: httpAddr.String(),
	}

	log.Fatal(server.ListenAndServe())
}
