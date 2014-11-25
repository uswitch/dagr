package program

import (
	"code.google.com/p/go-uuid/uuid"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type Execution struct {
	Program          *Program
	Id               string
	StartTime        time.Time
	recordedMessages []*executionMessage
	messages         chan *executionMessage
	finished         bool
	duration         time.Duration
	exitStatus       ExitCode
	subscribers      map[*websocket.Conn]bool
	sync.RWMutex
}

type executionMessage struct {
	ProgramName string `json:"programName"`
	MessageType string `json:"messageType"`
	Line        string `json:"line"`
}

func NewExecution(p *Program, messages chan *executionMessage) *Execution {
	e := &Execution{
		Program:     p,
		Id:          uuid.New(),
		StartTime:   time.Now(),
		messages:    messages,
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
	log.Println(e.Program.Name, messageType, message)
}

func (e *Execution) RecordedMessages() []*executionMessage {
	return e.recordedMessages
}

func (e *Execution) Finished() bool {
	return e.finished
}

func (e *Execution) ExitStatus() ExitCode {
	return e.exitStatus
}

func (e *Execution) Finish(exitStatus ExitCode) {
	e.Lock()
	defer e.Unlock()
	e.finished = true
	e.duration = time.Now().Sub(e.StartTime)
	e.exitStatus = exitStatus
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
	log.Println("adding subscriber")
	e.subscribers[c] = true
}

func (e *Execution) Unsubscribe(c *websocket.Conn) {
	log.Println("removing subscriber")
	delete(e.subscribers, c)
}

func (e *Execution) broadcast(msg *executionMessage) {
	e.RLock()
	defer e.RUnlock()
	for conn := range e.subscribers {
		if err := conn.WriteJSON(msg); err != nil {
			log.Println("error when sending to websocket", err)
		}
	}
}
