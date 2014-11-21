package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type Program struct {
	Name string `json:"name"`
	CommandPath string
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
