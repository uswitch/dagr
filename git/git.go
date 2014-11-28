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
	log.Println("cloning", repo, "into", workingDir)
	cmd := exec.Command("git", "clone", repo, workingDir)
	output, err := cmd.CombinedOutput()
	log.Println(string(output))
	
	return err
}

func Pull(workingDir string) error {
	log.Println("pulling latest into", workingDir)
	cmd := exec.Command("git", "pull")
	cmd.Dir = workingDir
	
	output, err := cmd.CombinedOutput()
	log.Println(string(output))
	
	return err
}

func MasterSha(repo string) (string, error) {
	log.Println("checking remote sha for", repo)	
	bytes, err := exec.Command("git", "ls-remote", repo, "master").CombinedOutput()
	parts := strings.Split(string(bytes), "\t")
	sha := parts[0]
	
	log.Println("latest remote sha:", sha)
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
