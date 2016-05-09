package web

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/uswitch/dagr/app"
	"github.com/uswitch/dagr/program"
	"log"
	"net/http"
	"strconv"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func programExecutions(app app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Println(fmt.Errorf("Can not upgrade to websocket: %v", err))
			return
		}
		vars := mux.Vars(req)
		programName := vars["program"]
		p := app.FindProgram(programName)
		if p == nil {
			log.Println("no such program:", programName)
			http.NotFound(w, req)
		} else {
			program.ProgramLog(p, "subscribing to executions")
			p.Subscribe(conn)
			go readLoop(p, conn)
		}
	}

}

func handleExecutionMessages(app app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Println(fmt.Errorf("Can not upgrade to websocket: %v", err))
			return
		}
		vars := mux.Vars(req)
		executionId := vars["executionId"]
		execution := app.FindExecution(executionId)
		if execution == nil {
			log.Println("no such execution:", executionId)
			http.NotFound(w, req)
		} else {
			program.ExecutionLog(execution, "subscribing to messages")
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

// read is required (http://www.gorillatoolkit.org/pkg/websocket)
func readLoop(s program.Subscriber, c *websocket.Conn) {
	for {
		_, _, err := c.NextReader()
		if err != nil {
			c.Close()
			s.Unsubscribe(c)
			return
		}
	}
}
