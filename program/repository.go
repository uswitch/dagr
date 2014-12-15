package program

import (
	"github.com/uswitch/dagr/git"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Repository struct {
	repo       string
	workingDir string
	sha        string
	programs   []*Program
}

func NewRepository(repo, workingDir string) (*Repository, error) {
	r := &Repository{repo, workingDir, "", nil}

	err := git.PullOrClone(repo, workingDir)

	if err != nil {
		return nil, err
	}

	err = r.refresh()

	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Repository) RunRefreshLoop(ticks <-chan time.Time) {
	for _ = range ticks {
		err := r.refresh()

		if err != nil {
			log.Println(err)
		}
	}
}

func (r *Repository) Programs() []*Program {
	return r.programs
}

func (r *Repository) FindProgram(name string) *Program {
	for _, program := range r.programs {
		if program.Name == name {
			return program
		}
	}

	return nil
}

func (r *Repository) refresh() error {
	newSha, err := git.MasterSha(r.repo)

	if err != nil {
		return err
	}

	if newSha == r.sha {
		return nil
	}

	err = git.Pull(r.workingDir)

	if err != nil {
		return err
	}

	programs, err := readDir(r.workingDir)

	if err != nil {
		return err
	}

	r.programs = programs
	r.sha = newSha

	return nil
}

func readDir(dir string) ([]*Program, error) {
	log.Println("looking for programs in", dir)
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	programs := []*Program{}

	for _, info := range infos {
		commandPath := filepath.Join(dir, info.Name(), "main")
		_, err := os.Stat(commandPath)

		if err == nil {
			mainSource, err := ioutil.ReadFile(commandPath)

			if err == nil {
				log.Println("program executable:", commandPath)

				configPath := filepath.Join(dir, info.Name(), "dagr.toml")
				config, err := readConfig(configPath)

				if err != nil {
					log.Println("invalid configuration file:", configPath)
					continue
				}

				p := newProgram(info.Name(), commandPath, string(mainSource), config)
				programs = append(programs, p)
			}
		}
	}

	return programs, nil
}
