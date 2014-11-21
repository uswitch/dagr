package main

import (
	"code.google.com/p/go-uuid/uuid"
	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"regexp"
	"text/template"
)

var TMPL = regexp.MustCompile(".tmpl$")

var resourceBox = rice.MustFindBox("resources")

type IndexPageState struct {
	Succeeded int
	Retryable int
	Failed    int
	Programs  []*Program
}

type InfoPageState struct {
	Program *Program
}

type ExecutionPageState struct {
	Execution    *Execution
	ExecutionUrl string
}

func handleIndex(dagr Dagr) func(http.ResponseWriter, *http.Request) {
	indexTemplate, err := loadTemplate("index.html.tmpl")

	if err != nil {
		log.Fatal(err)
	}

	return func(w http.ResponseWriter, req *http.Request) {
		if err := indexTemplate.Execute(w, IndexPageState{77, 13, 12, dagr.AllPrograms()}); err != nil {
			log.Println("error when executing index template:", err)
			http.Error(w, err.Error(), 500)
		}
	}
}

func handleInfo(dagr Dagr) func(http.ResponseWriter, *http.Request) {
	infoTemplate, err := loadTemplate("info.html.tmpl")

	if err != nil {
		log.Fatal(err)
	}

	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		programName := vars["program"]
		program := dagr.FindProgram(programName)
		if program == nil {
			log.Println("no such program:", programName)
			http.NotFound(w, req)
		} else if err := infoTemplate.Execute(w, InfoPageState{program}); err != nil {
			log.Println("error when executing info template:", err)
			http.Error(w, err.Error(), 500)
		}
	}
}

func handleExecution(dagr Dagr) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		programName := vars["program"]
		program := dagr.FindProgram(programName)
		if program == nil {
			log.Println("no such program:", programName)
			http.NotFound(w, req)
		} else {
			exec := NewExecution(program)
			guid := uuid.New()
			dagr.AddExecution(guid, exec)
			exec.Execute()
			http.Redirect(w, req, "/executions/"+guid, 302)
		}
	}
}

func showExecution(dagr Dagr) func(http.ResponseWriter, *http.Request) {
	showTemplate, err := loadTemplate("show.html.tmpl")

	if err != nil {
		log.Fatal(err)
	}

	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		executionId := vars["executionId"]
		execution := dagr.FindExecution(executionId)
		if execution == nil {
			log.Println("no such execution:", executionId)
			http.NotFound(w, req)
		} else {
			executionUrl := "ws://localhost:8080/executions/" + executionId + "/messages"

			if err := showTemplate.Execute(w, ExecutionPageState{execution, executionUrl}); err != nil {
				log.Println("error when executing execution template:", err)
				http.Error(w, err.Error(), 500)
			}
		}
	}
}

func loadTemplate(path string) (*template.Template, error) {
	templateString, err := resourceBox.String(path)
	if err != nil {
		return nil, err
	}
	var t = template.New(path)
	return t.Parse(templateString)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handleExecutionMessages(dagr Dagr) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Println("cannot upgrade to websocket")
			return
		}
		vars := mux.Vars(req)
		executionId := vars["executionId"]
		log.Println("broadcasting messages for execution id:", executionId)

		conn.WriteMessage(websocket.TextMessage, []byte("hello "+executionId))
	}
}

func Serve(httpAddr string, dagr Dagr) error {

	r := mux.NewRouter()
	r.HandleFunc("/", handleIndex(dagr)).Methods("GET")
	r.HandleFunc("/program/{program}", handleInfo(dagr)).Methods("GET")
	r.HandleFunc("/program/{program}/execute", handleExecution(dagr)).Methods("POST")
	r.HandleFunc("/executions/{executionId}", showExecution(dagr)).Methods("GET")
	r.HandleFunc("/executions/{executionId}/messages", handleExecutionMessages(dagr))
	http.Handle("/", r)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(resourceBox.HTTPBox())))

	server := &http.Server{
		Addr: httpAddr,
	}

	log.Println("dagr listening on", httpAddr)

	return server.ListenAndServe()
}
