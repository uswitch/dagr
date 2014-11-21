package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type Program struct {
	Name string `json:"name"`
	CommandPath string
}

type ExecutionWriter struct {
	ProgramName string
}

func NewExecutionWriter(p *Program) *ExecutionWriter {
	return &ExecutionWriter{p.Name}
}

func (e *ExecutionWriter) Write(bs []byte) (n int, err error) {
	s := string(bs[:])
	log.Println(e.ProgramName, ":", s)
	return len(bs), nil
}

func (p *Program) Execute() {
	log.Println("executing", p.CommandPath)
	cmd := exec.Command(p.CommandPath)
	w := NewExecutionWriter(p)
	cmd.Stdout = w
	cmd.Stderr = w
	
	err := cmd.Run()
	
	if err != nil {
		log.Println(err)
	}
	
	log.Println("finished executing", p.Name)
}

// does the given directory contain a 'main' file?
func isProgram(parentDir, dir string) bool {
	_, err := os.Stat(filepath.Join(parentDir, dir, "main"))
	return err == nil
}

func readDir(dir string) ([]*Program, error) {
	programs := []*Program{}

	log.Println("looking for programs in", dir)
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return programs, err
	}

	for _, info := range infos {
		if err == nil && info.IsDir() && isProgram(dir, info.Name()) {
			commandPath := filepath.Join(dir, info.Name(), "main")
			log.Println("program executable:", commandPath)

			programs = append(programs, &Program{info.Name(), commandPath})
		}
	}

	return programs, nil
}
