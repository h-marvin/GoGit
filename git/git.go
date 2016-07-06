package git

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	git = "git"
)

//IsRepository - Evaluates whether the given folder points to a git repository
func IsRepository(folderPath string) bool {
	children, _ := ioutil.ReadDir(folderPath)

	for _, c := range children {
		if c.IsDir() && strings.EqualFold(c.Name(), ".git") {
			return true
		}
	}

	return false
}

//SyncRepo - Determines based on .git/config whether the given git repository should be synchronized or not
func SyncRepo(repoPath string, filter string) bool {
	configPath := repoPath + "/.git/config"
	lines, err := readConfig(configPath)
	sync := false

	if err == nil {
		for i := range lines {
			if strings.Contains(lines[i], filter) {
				sync = true
				break
			}
		}
	} else {
		log.Fatal(err)
	}

	return sync
}

func readConfig(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

//Fetch - Will perform a 'git fetch --prune' command on the given repositry
func Fetch(repoPath string) error {
	cmd := git
	args := []string{"-C", repoPath, "fetch", "--prune"}

	return exec.Command(cmd, args...).Run()
}

//Pull - Will perform a 'git pull' command on the given repository
func Pull(repoPath string) error {
	cmd := git
	args := []string{"-C", repoPath, "pull"}

	return exec.Command(cmd, args...).Run()
}

//GetBranchName - Determines the name of the checked out branch of the given repository
func GetBranchName(repoPath string) string {
	cmd := git
	args := []string{"-C", repoPath, "name-rev", "--name-only", "HEAD"}

	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return "no branch name"
	}

	return strings.TrimSpace(string(out))
}
