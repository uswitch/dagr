package main

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
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

type ExecutionState struct {
	Execution            *Execution
	ExecutionSubscribers *ExecutionSubscribers
}

func (e *ExecutionState) Subscribe(c *websocket.Conn) {
	log.Println("adding subscriber")
	e.ExecutionSubscribers.Subscribe(c)
}
func (e *ExecutionState) Unsubscribe(c *websocket.Conn) {
	log.Println("removing subscriber")
	e.ExecutionSubscribers.Unsubscribe(c)
}
func (e *ExecutionState) StartRelay() {
	go func() {
		for msg := range e.Execution.Writer.Message {
			e.ExecutionSubscribers.BroadcastMessage(msg)
		}
	}()
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
			executionState := &ExecutionState{exec, NewExecutionSubscribers()}
			dagr.AddExecution(guid, executionState)

			exec.Execute()              // FIXME -- this may return an error
			executionState.StartRelay() // FIXME -- has to be run after Execute()

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
			executionUrl := fmt.Sprintf("/executions/%s/messages", executionId)
			log.Println("socket path:", executionUrl)
			// executionUrl := "ws://localhost:8080/executions/" + executionId + "/messages"

			if err := showTemplate.Execute(w, ExecutionPageState{execution.Execution, executionUrl}); err != nil {
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

type ExecutionSubscribers struct {
	subscribers map[*websocket.Conn]bool
}

func NewExecutionSubscribers() *ExecutionSubscribers {
	return &ExecutionSubscribers{make(map[*websocket.Conn]bool)}
}
func (e *ExecutionSubscribers) Unsubscribe(c *websocket.Conn) {
	delete(e.subscribers, c)
}
func (e *ExecutionSubscribers) Subscribe(c *websocket.Conn) {
	e.subscribers[c] = true
}
func (e *ExecutionSubscribers) BroadcastMessage(msg string) {
	for conn := range e.subscribers {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintln(msg)))
	}
}

// read is required (http://www.gorillatoolkit.org/pkg/websocket)
func readLoop(executionState *ExecutionState, c *websocket.Conn) {
	for {
		_, _, err := c.NextReader()
		if err != nil {
			c.Close()
			executionState.Unsubscribe(c)
			return
		}
	}
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
		executionState := dagr.FindExecution(executionId)

		executionState.Subscribe(conn)
		go readLoop(executionState, conn)
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
