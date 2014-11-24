package main

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

type Execution struct {
	Program          *Program
	Id               string
	recordedMessages []*ExecutionMessage
	messages         chan *ExecutionMessage
	exitStatus       chan ExitCode
	subscribers      map[*websocket.Conn]bool
	sync.RWMutex
}

type ExecutionMessage struct {
	ProgramName string `json:"programName"`
	MessageType string `json:"messageType"`
	Line        string `json:"line"`
}

func (e *Execution) sendMessage(messageType, message string) {
	e.Lock()
	defer e.Unlock()
	executionMessage := &ExecutionMessage{e.Program.Name, messageType, message + "\n"}
	e.messages <- executionMessage
	e.recordedMessages = append(e.recordedMessages, executionMessage)
	log.Println(e.Program.Name, messageType, message)
}

func (e *Execution) RecordedMessages() []*ExecutionMessage {
	e.RLock()
	defer e.RUnlock()
	return e.recordedMessages
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

func (e *Execution) Broadcast(msg *ExecutionMessage) {
	e.RLock()
	defer e.RUnlock()
	for conn := range e.subscribers {
		if err := conn.WriteJSON(msg); err != nil {
			log.Println("error when sending to websocket", err)
		}
	}
}

func (e *Execution) BroadcastAll() {
	for msg := range e.messages {
		e.Broadcast(msg)
	}
}
