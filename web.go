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

type ProgramStatus struct {
	Program           *Program
	LastExecution     *Execution
	LastExecutionTime string
	Running           bool
	Succeeded         bool
	Failed            bool
	Retryable         bool
}

type IndexPageState struct {
	Succeeded       int
	Retryable       int
	Failed          int
	ProgramStatuses []*ProgramStatus
}

type ProgramPageState struct {
	Program *Program
}

type ExecutionPageState struct {
	Execution *Execution
}

func handleIndex(dagr Dagr) http.HandlerFunc {
	indexTemplate := template.Must(loadTemplate("index.html.tmpl"))

	return func(w http.ResponseWriter, req *http.Request) {
		programs := dagr.AllPrograms()
		programStatuses := []*ProgramStatus{}

		var totalSucceeded, totalFailed, totalRetryable int

		for _, program := range programs {
			executions := program.Executions()
			var lastExecution *Execution
			var lastExecutionTime string
			if len(executions) > 0 {
				lastExecution = executions[len(executions)-1]
				lastExecutionTime = lastExecution.StartTime.Format("2 Jan 2006 15:04")
			}

			var running, succeeded, retryable, failed bool

			if lastExecution != nil {
				running = !lastExecution.Finished()

				if !running {
					succeeded = lastExecution.ExitStatus() == Success
					retryable = lastExecution.ExitStatus() == Retryable
					failed = lastExecution.ExitStatus() == Failed
				}
			}

			programStatuses = append(programStatuses,
				&ProgramStatus{
					Program:           program,
					LastExecution:     lastExecution,
					LastExecutionTime: lastExecutionTime,
					Running:           running,
					Succeeded:         succeeded,
					Retryable:         retryable,
					Failed:            failed,
				})

			if succeeded {
				totalSucceeded++
			}

			if retryable {
				totalRetryable++
			}

			if failed {
				totalFailed++
			}
		}

		err := indexTemplate.Execute(w, IndexPageState{totalSucceeded, totalRetryable, totalFailed, programStatuses})

		if err != nil {
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
		} else if err := infoTemplate.Execute(w, ProgramPageState{program}); err != nil {
			log.Println("error when executing program info template:", err)
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
		} else if err := showTemplate.Execute(w, ExecutionPageState{execution}); err != nil {
			log.Println("error when executing execution info template:", err)
			http.Error(w, err.Error(), 500)
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
		if execution == nil {
			log.Println("no such execution:", executionId)
			http.NotFound(w, req)
		} else {
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
