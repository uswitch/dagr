package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

type Program struct {
	Dir  string `json:"dir"`
	Name string `json:"name"`
}

func (*Program) Execute() *exec.Cmd {
	return &exec.Cmd{}
}

// does the given directory contain a 'main' file?
func isProgram(parentDir, dir string) bool {
	_, err := os.Stat(filepath.Join(parentDir, dir, "main"))
	return err == nil
}

func readDir(dir string) ([]*Program, error) {
	programs := []*Program{}

	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return programs, err
	}

	for _, info := range infos {
		if err == nil && info.IsDir() && isProgram(dir, info.Name()) {
			programs = append(programs, &Program{dir, info.Name()})
		}
	}
	return programs, nil
}
