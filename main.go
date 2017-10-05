// Use of this source code is governed by a
// license that can be found in the LICENSE file.
package main

import (
	"flag"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/h-marvin/GoGit/git"
)

type command int

const (
	// CLEAN - to perform a 'git gc' command
	CLEAN command = iota
	// FETCH - to perform a 'git fetch --prune' command
	FETCH
	// PULL - to perform a 'git pull' command
	PULL
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
	fetch := flag.Bool("fetch", false, "Decide whether you want to perform a 'fetch --prune' request instead of a pull request.")
	clean := flag.Bool("clean", false, "Decide whether you want to perform a 'gc' request instead of a pull request.")
	flag.Parse()

	var action command
	if *clean {
		action = CLEAN
	} else if *fetch {
		action = FETCH
	} else {
		action = PULL
	}

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

	c := make(chan string, 4)
	var wg sync.WaitGroup
	var repoCount int8
	wg.Add(1)
	timeout := time.After(60 * time.Second)

	go func() {
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
						wg.Add(1)
						repoCount++
						go func() {
							defer wg.Done()
							c <- performGitCommands(path, gitRoot, action)
						}()
					}
				}
				return nil
			})
		}
		wg.Done()
	}()

	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	for {
		select {
		case result := <-c:
			log.Info(result)
		case <-done:
			log.Infof("All %d repos have been synced.", repoCount)
			return
		case <-timeout:
			log.Info("Operation timed out.")
			return
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

func performGitCommands(repoPath string, gitRoot string, action command) string {
	branchDesc := getRepoName(repoPath) + " | "

	branchName := git.GetBranchName(repoPath)
	if printBranchName(branchName) {
		branchDesc += branch + "  " + branchName + " | "
	}

	if action == FETCH {
		return "fetch | " + branchDesc + printResult(git.Fetch(repoPath))
	}
	if action == CLEAN {
		return "clean | " + branchDesc + printResult(git.Clean(repoPath))
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
