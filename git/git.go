package git

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

// does the given directory contain a '.git' directory?
func IsRepo(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}

func Clone(repo, workingDir string) error {
	return exec.Command("git", "clone", repo, workingDir).Run()
}

func Pull(workingDir string) error {
	return exec.Command("git", "-C", workingDir, "pull").Run()
}

func MasterSha(workingDir string) (string, error) {
	bytes, err := ioutil.ReadFile(filepath.Join(workingDir, ".git", "refs", "heads", "master"))
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// performs Clone or Pull as necessary
func Update(repo, workingDir string) error {
	if !IsRepo(workingDir) {
		return Clone(repo, workingDir)
	} else {
		return Pull(workingDir)
	}
}
