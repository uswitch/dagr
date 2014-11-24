package main

import (
	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
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
	Execution *Execution
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

func handleProgramExecute(dagr Dagr) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		programName := vars["program"]
		program := dagr.FindProgram(programName)
		if program == nil {
			log.Println("no such program:", programName)
			http.NotFound(w, req)
		} else {
			execution, err := dagr.Execute(program)

			if err != nil {
				log.Println("error on execution:", err)
				http.Error(w, err.Error(), 500)
				return
			}

			http.Redirect(w, req, "/executions/"+execution.Id, 302)
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
			if err := showTemplate.Execute(w, ExecutionPageState{execution}); err != nil {
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
		log.Println("subscribing to messages for execution id:", executionId)
		execution := dagr.FindExecution(executionId)
		execution.Subscribe(conn)
		countSoFarStr := vars["countSoFar"]
		countSoFar, err := strconv.Atoi(countSoFarStr)
		if err != nil {
			log.Println("countSoFar not an integer?", countSoFarStr, err)
		} else {
			messagesCaughtUp := execution.CatchUp(conn, countSoFar)
			if messagesCaughtUp > 0 {
				log.Println("caught up", messagesCaughtUp, "message(s)")
			}
		}

		go readLoop(execution, conn)
	}
}

func DagrHandler(dagr Dagr) http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", handleIndex(dagr)).Methods("GET")
	r.HandleFunc("/program/{program}", handleProgramInfo(dagr)).Methods("GET")
	r.HandleFunc("/program/{program}/execute", handleProgramExecute(dagr)).Methods("POST")
	r.HandleFunc("/executions/{executionId}", handleExecutionInfo(dagr)).Methods("GET")
	r.HandleFunc("/executions/{executionId}/messages/{countSoFar:[0-9]+}", handleExecutionMessages(dagr))
	return r
}
