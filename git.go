package main

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

func MasterSha(repo string) (string, error) {
	bytes, err := exec.Command("git", "ls-remote", repo, "master").Output()

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// performs Pull or Clone as necessary
func PullOrClone(repo, workingDir string) error {
	if IsRepo(workingDir) {
		return Pull(workingDir)
	} else {
		return Clone(repo, workingDir)
	}
}
