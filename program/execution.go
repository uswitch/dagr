package program

import (
	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type ExitCode int

const (
	SuccessCode   = 0
	RetryableCode = 1
	FailedCode    = 2
)

type Status struct {
	name, label string
}

var SuccessStatus = &Status{"succeeded", "Succeeded"}
var RetryableStatus = &Status{"retryable", "Retryable"}
var FailedStatus = &Status{"failed", "Failed"}
var RunningStatus = &Status{"running", "Running"}

type Execution struct {
	Program          *Program
	Id               string
	StartTime        time.Time
	recordedMessages []*executionMessage
	messages         chan *executionMessage
	finished         bool
	duration         time.Duration
	exitCode         ExitCode
	subscribers      map[*websocket.Conn]bool
	sync.RWMutex
}

type executionMessage struct {
	ProgramName string `json:"programName"`
	MessageType string `json:"messageType"`
	Line        string `json:"line"`
}

func NewExecution(p *Program) *Execution {
	e := &Execution{
		Program:     p,
		Id:          uuid.New(),
		StartTime:   time.Now(),
		messages:    make(chan *executionMessage, BUFFER_SIZE),
		subscribers: make(map[*websocket.Conn]bool),
	}
	go func() {
		for msg := range e.messages {
			e.broadcast(msg)
		}
	}()

	return e
}

func (e *Execution) SendMessage(messageType, message string) {
	e.Lock()
	defer e.Unlock()
	executionMessage := &executionMessage{e.Program.Name, messageType, message + "\n"}
	e.messages <- executionMessage
	e.recordedMessages = append(e.recordedMessages, executionMessage)
	//log.Println(e.Program.Name, messageType, message)
}

func (e *Execution) RecordedMessages() []*executionMessage {
	return e.recordedMessages
}

func (e *Execution) Finished() bool {
	return e.finished
}

func (e *Execution) ExitCode() ExitCode {
	return e.exitCode
}

func (e *Execution) Finish(exitCode ExitCode) {
	e.Lock()
	defer e.Unlock()
	e.finished = true
	e.duration = time.Now().Sub(e.StartTime)
	e.exitCode = exitCode
	e.Program.SendExecutionState(e)
}

func (e *Execution) Duration() time.Duration {
	return e.duration
}

func (e *Execution) CatchUp(conn *websocket.Conn, countSoFar int) int {
	e.RLock()
	defer e.RUnlock()

	messages := e.recordedMessages[countSoFar:]

	for _, msg := range messages {
		if err := conn.WriteJSON(msg); err != nil {
			log.Println("error when sending to websocket (catch up)", err)
		}
	}

	return len(messages)
}

func (e *Execution) Subscribe(c *websocket.Conn) {
	ExecutionLog(e, "adding subscriber")
	e.subscribers[c] = true
}

func (e *Execution) Unsubscribe(c *websocket.Conn) {
	ExecutionLog(e, "removing subscriber")
	delete(e.subscribers, c)
}

func (e *Execution) broadcast(msg *executionMessage) {
	e.RLock()
	defer e.RUnlock()
	for conn := range e.subscribers {
		if err := conn.WriteJSON(msg); err != nil {
			ExecutionLog(e, "error when sending to websocket", err)
		}
	}
}

func (e *Execution) Status() *Status {
	if e.finished {
		switch e.exitCode {
		case SuccessCode:
			return SuccessStatus
		case RetryableCode:
			return RetryableStatus
		default:
			return FailedStatus
		}
	}
	return RunningStatus
}
