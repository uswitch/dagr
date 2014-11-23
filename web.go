package main

import (
	"fmt"
	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"text/template"
)

var templatesBox = rice.MustFindBox("resources/templates")

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
	Program      *Program
	ExecutionUrl string
}

func handleIndex(dagr Dagr) http.HandlerFunc {
	indexTemplate := template.Must(loadTemplate("index.html.tmpl"))

	return func(w http.ResponseWriter, req *http.Request) {
		if err := indexTemplate.Execute(w, IndexPageState{77, 13, 12, dagr.AllPrograms()}); err != nil {
			log.Println("error when executing index template:", err)
			http.Error(w, err.Error(), 500)
		}
	}
}

func handleProgramInfo(dagr Dagr) http.HandlerFunc {
	infoTemplate := template.Must(loadTemplate("program.html.tmpl"))

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

type Execution struct {
	id          string
	program     *Program
	subscribers map[*websocket.Conn]bool
	sync.RWMutex
}

func (e *Execution) Subscribe(c *websocket.Conn) {
	e.Lock()
	defer e.Unlock()
	log.Println("adding subscriber")
	e.subscribers[c] = true
}
func (e *Execution) Unsubscribe(c *websocket.Conn) {
	e.Lock()
	defer e.Unlock()
	log.Println("removing subscriber")
	delete(e.subscribers, c)
}
func (e *Execution) Broadcast(msg string) {
	e.RLock()
	defer e.RUnlock()
	for conn := range e.subscribers {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintln(msg)))
	}
}
func (e *Execution) BroadcastAllMessages(messages chan string) {
	for msg := range messages {
		e.Broadcast(msg)
	}
}

func handleProgramExecute(dagr Dagr) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		programName := vars["program"]
		program := dagr.FindProgram(programName)
		if program == nil {
			log.Println("no such program:", programName)
			http.NotFound(w, req)
		} else {
			execution := dagr.AddExecution(program)
			executionResult, err := program.Execute()

			if err != nil {
				log.Println("error on execution:", err)
				http.Error(w, err.Error(), 500)
				return
			}

			go execution.BroadcastAllMessages(executionResult.Stdout)

			http.Redirect(w, req, "/executions/"+execution.id, 302)
		}
	}
}

func handleExecutionInfo(dagr Dagr) http.HandlerFunc {
	showTemplate := template.Must(loadTemplate("execution.html.tmpl"))

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

			if err := showTemplate.Execute(w, ExecutionPageState{execution.program, executionUrl}); err != nil {
				log.Println("error when executing execution template:", err)
				http.Error(w, err.Error(), 500)
			}
		}
	}
}

func loadTemplate(path string) (*template.Template, error) {
	templateString, err := templatesBox.String(path)
	if err != nil {
		return nil, err
	}
	return template.New(path).Parse(templateString)
}

// read is required (http://www.gorillatoolkit.org/pkg/websocket)
func readLoop(execution *Execution, c *websocket.Conn) {
	for {
		_, _, err := c.NextReader()
		if err != nil {
			c.Close()
			execution.Unsubscribe(c)
			return
		}
	}
}

func handleExecutionMessages(dagr Dagr) http.HandlerFunc {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	return func(w http.ResponseWriter, req *http.Request) {
		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Println("cannot upgrade to websocket")
			return
		}
		vars := mux.Vars(req)
		executionId := vars["executionId"]
		log.Println("broadcasting messages for execution id:", executionId)
		execution := dagr.FindExecution(executionId)

		execution.Subscribe(conn)
		go readLoop(execution, conn)
	}
}

func DagrHandler(dagr Dagr) http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", handleIndex(dagr)).Methods("GET")
	r.HandleFunc("/program/{program}", handleProgramInfo(dagr)).Methods("GET")
	r.HandleFunc("/program/{program}/execute", handleProgramExecute(dagr)).Methods("POST")
	r.HandleFunc("/executions/{executionId}", handleExecutionInfo(dagr)).Methods("GET")
	r.HandleFunc("/executions/{executionId}/messages", handleExecutionMessages(dagr))
	return r
}
