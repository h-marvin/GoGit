// Use of this source code is governed by a
// license that can be found in the LICENSE file.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/h-marvin/GoGit/git"
)

const (
	branch        = "🌿️"
	pathSeparator = ":"
	failed        = "💥"
	success       = "👍🏼️"
)

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
		gitRoot = ensureTrailingSlash(gitRoot)
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

			if git.IsRepository(path) {
				rememberRepo(path, gitLocations)
				if len(*filter) == 0 || git.SyncRepo(path, *filter) {
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

func ensureTrailingSlash(path string) string {
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

func rememberRepo(path string, gitLocations map[string]int) {
	gitLocations[repoKey(path)] = 1
}

func repoKey(path string) string {
	return path + "/"
}

func performGitCommands(repoPath string, gitRoot string, fetch bool) string {
	branchDesc := getRepoName(repoPath) + " | "

	branchName := git.GetBranchName(repoPath)
	if printBranchName(branchName) {
		branchDesc += branch + "  " + branchName + " | "
	}

	if fetch {
		return "fetch | " + branchDesc + printResult(git.Fetch(repoPath))
	}
	return branchDesc + printResult(git.Pull(repoPath))
}

func printBranchName(name string) bool {
	isMaster := strings.EqualFold(name, "master")
	isTag := strings.HasPrefix(name, "tags/")

	return !isMaster && !isTag
}

func getRepoName(repoPath string) string {
	pathElements := strings.Split(repoPath, "/")
	return strings.TrimSuffix(pathElements[len(pathElements)-1], ".git")
}

func printResult(err error) string {
	if err != nil {
		return failed
	}
	return success
}
