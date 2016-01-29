// Use of this source code is governed by a
// license that can be found in the LICENSE file.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

const failed = "💥"
const success = "👍🏼️"
const branch = "🌿️"
const pathSeparator = ":"

func main() {
	inputGitRoots := flag.String("path", "", "The path to the root folder to start looking for git repositories. More than one location can be passed with a colon as separator.")
	filter := flag.String("filter", "", "Allows filtering for a certain value to occur in the .git/config (e.g. enterprise git address).")
	recursive := flag.Bool("recursive", false, "If 'false' only the first level of folders will be checked. If 'true' sub folders will be checked also (will not check within sub folders of git repos).")
	fetch := flag.Bool("fetch", false, "Decide whether you want to perform a fetch --prune request instead of a pull request.")
	flag.Parse()

	var gitRoots []string
	if len(*inputGitRoots) == 0 {
		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		gitRoots = append(gitRoots, usr.HomeDir)
	} else {
		gitRoots = strings.Split(*inputGitRoots, pathSeparator)
	}

	updates := 0
	c := make(chan string)
	timeout := time.After(30 * time.Second)

	for _, gitRoot := range gitRoots {
		gitRoot = trailingSlash(gitRoot)
		gitLocations := make(map[string]int)
		filepath.Walk(gitRoot, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				return nil
			}

			if strings.EqualFold(path, gitRoot) {
				return nil
			}

			if !*recursive && strings.Contains(trimPath(path, gitRoot), "/") {
				return nil
			}

			if withinGitRepo(path, gitLocations) {
				return nil
			}

			if isGitRepo(path) {
				rememberRepo(path, gitLocations)
				if len(*filter) == 0 || syncGitRepo(path, *filter) {
					updates++
					go func() { c <- performGitCommands(path, gitRoot, *fetch) }()
				}
			}

			return nil
		})
	}

	for i := 0; i < updates; i++ {
		select {
		case result := <-c:
			fmt.Println(result)
		case <-timeout:
			fmt.Println("timedout")
		}
	}
}

func trailingSlash(path string) string {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}

func trimPath(path string, gitRoot string) string {
	return strings.TrimPrefix(path, gitRoot)
}

func withinGitRepo(path string, gitLocations map[string]int) bool {
	for location := range gitLocations {
		if strings.HasPrefix(path, location) {
			return true
		}
	}
	return false
}

func isGitRepo(folderPath string) bool {
	children, _ := ioutil.ReadDir(folderPath)

	for _, c := range children {
		if c.IsDir() && strings.EqualFold(c.Name(), ".git") {
			return true
		}
	}

	return false
}

func rememberRepo(path string, gitLocations map[string]int) {
	gitLocations[repoKey(path)] = 1
}

func repoKey(path string) string {
	return path + "/"
}

func syncGitRepo(repoPath string, filter string) bool {
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

func performGitCommands(repoPath string, gitRoot string, fetch bool) string {
	repo := trimPath(repoPath, gitRoot)
	branchDesc := repo + " | "

	branchName := getBranchName(repoPath)
	if !strings.EqualFold(branchName, "master") {
		branchDesc += branch + "  " + branchName + " | "
	}

	if fetch {
		return "fetch | " + branchDesc + performGitFetch(repoPath)
	}
	return branchDesc + performGitPull(repoPath)
}

func getBranchName(repoPath string) string {
	cmd := "git"
	args := []string{"-C", repoPath, "name-rev", "--name-only", "HEAD"}

	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return "no branch name"
	}

	return strings.TrimSpace(string(out))
}

func performGitFetch(repoPath string) string {
	cmd := "git"
	args := []string{"-C", repoPath, "fetch", "--prune"}

	err := exec.Command(cmd, args...).Run()
	if err != nil {
		return failed
	}

	return success
}

func performGitPull(repoPath string) string {
	cmd := "git"
	args := []string{"-C", repoPath, "pull"}

	err := exec.Command(cmd, args...).Run()
	if err != nil {
		return failed
	}

	return success
}
