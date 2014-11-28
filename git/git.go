package git

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// does the given directory contain a '.git' directory?
func IsRepo(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}

func Clone(repo, workingDir string) error {
	cmd := exec.Command("git", "clone", repo, workingDir)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		log.Println("error cloning repository", string(output), err)
	}
	
	return err
}

func Pull(workingDir string) error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = workingDir
	
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		log.Println("error pulling latest into", workingDir, err)
		log.Println(output)
	}
	
	return err
}

func MasterSha(repo string) (string, error) {
	bytes, err := exec.Command("git", "ls-remote", repo, "master").CombinedOutput()

	if err != nil {
		log.Println("error getting remote sha for", repo, err)
		return "", err
	}

	parts := strings.Split(string(bytes), "\t")
	sha := parts[0]

	return sha, nil
}

// performs Pull or Clone as necessary
func PullOrClone(repo, workingDir string) error {
	log.Println(repo, "initialising into", workingDir)
	if IsRepo(workingDir) {
		return Pull(workingDir)
	} else {
		return Clone(repo, workingDir)
	}
}
