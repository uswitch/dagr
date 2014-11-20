package web

import (
	"bitbucket.org/tebeka/nrsc"
	"github.com/uswitch/dagr/program"
	"net/http"
	"regexp"
)

var TMPL = regexp.MustCompile(".tmpl$")

func handleIndex(programs []*program.Program) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		t, err := nrsc.LoadTemplates(nil, "index.html.tmpl")
		if err != nil {
			http.NotFound(w, req)
		}
		if err = t.Execute(w, program.Status{77, 13, 12, programs}); err != nil {
			http.NotFound(w, req)
		}
	}
}

func Serve(httpAddr string, programs []*program.Program) error {
	nrsc.Handle("/static/")
	nrsc.Mask(TMPL)
	http.HandleFunc("/", handleIndex(programs))

	server := &http.Server{
		Addr: httpAddr,
	}

	return server.ListenAndServe()
}
