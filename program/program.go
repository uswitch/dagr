package program

import (
	"bufio"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
)

const BUFFER_SIZE = 1000

type Subscriber interface {
	Unsubscribe(*websocket.Conn)
}

type Program struct {
	Name        string
	CommandPath string
	MainSource  string
	executions  []*Execution
	messages    chan *programExecutionsMessage
	subscribers map[*websocket.Conn]bool
	Config
	sync.RWMutex
}

type programExecutionsMessage struct {
	ProgramName          string `json:"programName"`
	ExecutionId          string `json:"executionId"`
	ExecutionTime        string `json:"executionTime"`
	ExecutionLastOutput  string `json:"executionLastOutput"`
	ExecutionStatus      string `json:"executionStatus"`
	ExecutionStatusLabel string `json:"executionStatusLabel"`
}

func forwardOutput(execution *Execution, messageType string, r io.Reader, finished chan interface{}) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		execution.SendMessage(messageType, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Println(execution.Program.Name, "scanner error", err)
	}

	finished <- struct{}{}
}

func (p *Program) Execute(startCh <-chan bool, ch chan<- ExitCode) (*Execution, error) {
	p.Lock()
	defer p.Unlock()

	ProgramLog(p, "executing command")

	cmd := exec.Command(p.CommandPath)
	cmd.Dir = filepath.Dir(p.CommandPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	execution := NewExecution(p, cmd)
	cmd.Env = append(cmd.Env, "DAGR_EXECUTION_ID="+execution.Id)

	p.SendExecutionState(execution)
	messages := execution.messages
	stdoutFinished := make(chan interface{})
	stderrFinished := make(chan interface{})

	go forwardOutput(execution, "out", stdout, stdoutFinished)
	go forwardOutput(execution, "err", stderr, stderrFinished)

	go func() {
		ExecutionLog(execution, "waiting for execution start signal")
		<-startCh

		err = cmd.Start()
		if err != nil {
			execution.SendMessage("fail", fmt.Sprintf("failed to start: %s", err.Error()))
			execution.Finish(FailedCode)
			ch <- FailedCode
			return
		}
		execution.Started()

		defer close(messages)
		defer close(stdoutFinished)
		defer close(stderrFinished)

		// docs say we shouldn't call cmd.Wait() until all has been read, hence
		// the need for the 'finished' channels
		<-stdoutFinished
		<-stderrFinished

		err := cmd.Wait()
		if err == nil {
			execution.SendMessage("ok", "successfully completed")
			execution.Finish(SuccessCode)
			ch <- SuccessCode
			return
		}

		exitCode, err := extractExitCode(err)

		if err != nil {
			log.Println(p.Name, "failed to run", err)
			execution.SendMessage("fail", fmt.Sprint("failed to run ", err))
			execution.Finish(FailedCode)
			ch <- FailedCode
			return
		}

		ExecutionLog(execution, "exited with status", exitCode)
		execution.SendMessage("fail", fmt.Sprint("exited with status ", exitCode))
		execution.Finish(exitCode)
		ch <- exitCode
	}()

	p.executions = append(p.executions, execution)

	return execution, nil
}

func (p *Program) Executions() []*Execution {
	return p.executions
}

func (p *Program) SendExecutionState(e *Execution) {
	status := e.Status()
	programExecutionsMessage := &programExecutionsMessage{
		p.Name,
		e.Id,
		e.StartTime.Format("2 Jan 2006 15:04"),
		e.LastOutput("out"),
		status.name,
		status.label,
	}
	p.messages <- programExecutionsMessage
}

func (p *Program) Subscribe(c *websocket.Conn) {
	ProgramLog(p, "adding subscriber")
	p.subscribers[c] = true
}

func (p *Program) Unsubscribe(c *websocket.Conn) {
	ProgramLog(p, "removing subscriber")
	delete(p.subscribers, c)
}

func (p *Program) broadcast(msg *programExecutionsMessage) {
	p.RLock()
	defer p.RUnlock()
	for conn := range p.subscribers {
		if err := conn.WriteJSON(msg); err != nil {
			ProgramLog(p, "error when sending to websocket", err)
		}
	}
}

func extractExitCode(err error) (ExitCode, error) {
	switch ex := err.(type) {
	case *exec.ExitError:
		return ExitCode(ex.Sys().(syscall.WaitStatus).ExitStatus()), nil // assume Unix
	default:
		return 0, err
	}
}

func newProgram(name, commandPath, mainSource string, config *Config) *Program {
	return &Program{
		Name:        name,
		CommandPath: commandPath,
		MainSource:  mainSource,
		Config:      *config,
		messages:    make(chan *programExecutionsMessage, BUFFER_SIZE),
		subscribers: make(map[*websocket.Conn]bool),
	}
}

func startBroadcasting(program *Program) {
	go func() {
		for msg := range program.messages {
			program.broadcast(msg)
		}
	}()
}

func update(existingProgram, newProgram *Program) {
	existingProgram.MainSource = newProgram.MainSource
	existingProgram.Config = newProgram.Config
}

func ProgramLog(p *Program, args ...interface{}) {
	_, fn, line, _ := runtime.Caller(1)
	identity := []string{p.Name}
	s := fmt.Sprintf("%-25s", fmt.Sprintf("%s:%d", filepath.Base(fn), line))
	log.Println(append([]interface{}{s, identity}, args...)...)
}

func ExecutionLog(e *Execution, args ...interface{}) {
	_, fn, line, _ := runtime.Caller(1)
	identity := []string{e.Program.Name, e.Id}
	s := fmt.Sprintf("%-25s", fmt.Sprintf("%s:%d", filepath.Base(fn), line))
	log.Println(append([]interface{}{s, identity}, args...)...)
}
