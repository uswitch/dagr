package program

import (
	"github.com/pborman/uuid"
	"github.com/gorilla/websocket"
	"log"
	"errors"
	"fmt"
	"net/http"
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
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
var WaitingStatus = &Status{"waiting", "Waiting"}
var slackWebhookUrl = os.Getenv("DAGR_SLACK_WEBHOOK_URL")
var slackAlertsRoom = os.Getenv("DAGR_SLACK_ALERTS_ROOM")
var slackSilenceRetryable = os.Getenv("DAGR_SLACK_SILENCE_RETRYABLE")
var slackUserName = os.Getenv("DAGR_SLACK_USER_NAME")
var slackIconEmoji = os.Getenv("DAGR_SLACK_ICON_EMOJI")
var slackIconUrl = os.Getenv("DAGR_SLACK_ICON_URL")
var slackNoWarnLevel = SuccessCode

func init() {
	if slackIconEmoji == "" {
		slackIconEmoji = ":exclamation:"
	}
	if slackUserName == "" {
		slackUserName = "Dagr"
	}
	if slackSilenceRetryable != "" {
		slackNoWarnLevel = RetryableCode
	}
}

type Execution struct {
	Program          *Program
	Id               string
	StartTime        time.Time
	cmd              *exec.Cmd
	recordedMessages []*executionMessage
	messages         chan *executionMessage
	started          bool
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

type slackMessage struct {
	Text        string        `json:"text"`
	Channel     string        `json:"channel,omitempty"`
	UserName    string        `json:"username,omitempty"`
	IconURL     string        `json:"icon_url,omitempty"`
	IconEmoji   string        `json:"icon_emoji,omitempty"`
}

func NewExecution(p *Program, cmd *exec.Cmd) *Execution {
	e := &Execution{
		Program:     p,
		Id:          uuid.New(),
		StartTime:   time.Now(),
		cmd:         cmd,
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

// sends a SIGINT to the running process
func (e *Execution) Shutdown() {
	if !e.IsRunning() {
		return
	}

	ExecutionLog(e, "sending SIGINT")
	e.cmd.Process.Signal(os.Interrupt)
	for {
		if e.IsRunning() {
			ExecutionLog(e, "waiting for shutdown...")
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
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

func (e *Execution) LastOutput(messageType string) string {
	for i := len(e.recordedMessages) - 1; i >= 0; i-- {
		if e.recordedMessages[i].MessageType == messageType {
			return strings.TrimSpace(e.recordedMessages[i].Line)
		}
	}
	return ""
}

func (e *Execution) IsRunning() bool {
	return e.started && !e.finished
}

func (e *Execution) Finished() bool {
	return e.finished
}

func (e *Execution) ExitCode() ExitCode {
	return e.exitCode
}

func (e *Execution) Started() {
	e.Lock()
	defer e.Unlock()
	e.started = true
}

func (e *Execution) slackWarningMaybe() error {
	if slackWebhookUrl != "" && slackAlertsRoom != "" && int(e.exitCode) > slackNoWarnLevel {
		msgText := fmt.Sprintf("`%s` exited with %s.\nLast output:\n```\n%s\n```", e.Program.Name, e.Status().label, e.LastOutput("out"))
		msg := &slackMessage{msgText, slackAlertsRoom, slackUserName, slackIconUrl, slackIconEmoji}
		buf, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		resp, err := http.Post(slackWebhookUrl, "application/json", bytes.NewReader(buf))
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			errors.New(fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
		}
	}
	return nil
}

func (e *Execution) Finish(exitCode ExitCode) {
	e.Lock()
	defer e.Unlock()
	e.finished = true
	e.duration = time.Now().Sub(e.StartTime)
	e.exitCode = exitCode
	err := e.slackWarningMaybe()
	if err != nil {
		log.Println("error when sending warning to slack", err)
	}
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

	if e.IsRunning() {
		return RunningStatus
	}

	return WaitingStatus
}
