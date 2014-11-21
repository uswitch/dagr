package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type Program struct {
	Name string `json:"name"`
}

type Status struct {
	Succeeded int
	Retryable int
	Failed    int
	Programs  []*Program
}

// does the given directory contain a 'main' file?
func IsProgram(parentDir, dir string) bool {
	_, err := os.Stat(filepath.Join(parentDir, dir, "main"))
	return err == nil
}

func ReadDir(dir string) ([]*Program, error) {
	programs := []*Program{}

	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return programs, err
	}

	for _, info := range infos {
		if err == nil && info.IsDir() && IsProgram(dir, info.Name()) {
			programs = append(programs, &Program{info.Name()})
		}
	}
	return programs, nil
}
