package main

import (
	"log"
	"os/exec"
	"syscall"
)

type ExecutionWriter struct {
	ProgramName string
}

func NewExecutionWriter(e *Execution) *ExecutionWriter {
	return &ExecutionWriter{e.Program.Name}
}

func (e *ExecutionWriter) Write(bs []byte) (n int, err error) {
	s := string(bs[:])
	log.Println(e.ProgramName, ":", s)
	return len(bs), nil
}

const (
	Success = 0
	Retryable = 1
	Failed = 2
)

type Execution struct {
	Program *Program
}

func (e *Execution) Execute() {
	log.Println("executing", e.Program.CommandPath)
	cmd := exec.Command(e.Program.CommandPath)

	w := NewExecutionWriter(e)
	cmd.Stdout = w
	cmd.Stderr = w
	
	err := cmd.Run()
	
	if err == nil {
		log.Println("finished executing", e.Program.Name)
	} else {
		log.Println("command error", err)
	
		executionError := err.(*exec.ExitError)
	
		if executionError != nil {
			ws := executionError.Sys().(syscall.WaitStatus)
			exitCode := ws.ExitStatus()
			log.Println("exit code", exitCode)
		}
	}
}

func NewExecution(program *Program) *Execution {
	return &Execution{program}
}
