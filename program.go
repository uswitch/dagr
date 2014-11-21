package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Program struct {
	Name string `json:"name"`
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
			programName := info.Name()
			log.Println("found program:", programName)
			
			programs = append(programs, &Program{info.Name()})
		}
	}
	
	return programs, nil
}

func MonitorPrograms(repo, workingDir string, delay time.Duration) (chan []*Program, error) {
	ch := make(chan []*Program)

	sha := ""

	go func() {
		for {
			defer func() {
				time.Sleep(delay)
			}()

			newSha, err := MasterSha(repo)

			if err != nil {
				log.Print(err)
				continue
			}

			if newSha != sha {
				log.Println("pulling from repository", repo)
				err := Pull(workingDir)

				if err != nil {
					log.Print(err)
					continue
				}

				newPrograms, err := readDir(workingDir)

				if err != nil {
					log.Print(err)
					continue
				}

				ch <- newPrograms
				sha = newSha
			}
		}
	}()

	return ch, nil
}

func FindProgram(name string, programs []*Program) *Program {
	for _, program := range programs {
		if program.Name == name {
			return program
		}
	}

	return nil
}
