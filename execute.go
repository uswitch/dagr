package main

import (
	"bufio"
	"io"
	"log"
	"os/exec"
)

const BUFFER_SIZE = 1000

type ExecutionWriter struct {
	ProgramName string
	Message     chan string
	stdout      io.ReadCloser
}

func NewExecutionWriter(e *Execution, cmd *exec.Cmd) (*ExecutionWriter, error) {
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return nil, err
	}

	return &ExecutionWriter{e.Program.Name, make(chan string, BUFFER_SIZE), stdout}, nil
}

func (e *ExecutionWriter) Copy() error {
	scanner := bufio.NewScanner(e.stdout)

	for scanner.Scan() {
		s := scanner.Text()
		log.Println(e.ProgramName, s)
		e.Message <- s
	}

	return scanner.Err()
}

const (
	Success   = 0
	Retryable = 1
	Failed    = 2
)

type Execution struct {
	Writer  *ExecutionWriter
	Program *Program
}

func (e *Execution) Execute() error {
	log.Println("executing", e.Program.CommandPath)
	cmd := exec.Command(e.Program.CommandPath)

	w, err := NewExecutionWriter(e, cmd)
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	e.Writer = w

	// go func() {
	//	log.Println("waiting to finish", e.Program.Name)
	//	cmd.Wait()
	//	log.Println("finished", e.Program.Name)
	// }()

	go w.Copy()

	return nil
}

func NewExecution(program *Program) *Execution {
	return &Execution{nil, program}
}
