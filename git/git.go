package git

import (
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

// performs Clone or Pull as necessary
func Update(repo, workingDir string) error {
	if !IsRepo(workingDir) {
		return Clone(repo, workingDir)
	} else {
		return Pull(workingDir)
	}
}
