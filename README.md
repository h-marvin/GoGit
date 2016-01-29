# GoGit
GoGit is a simple to use Go script allowing you to conveniently update all your local git repositories with the latest changes. You decide whether you want fetch or pull. Making use of Go's channels, found repositories will be refreshed in a non-blocking manner, which allows big amounts of repositories to be updated in no time.

## Installation
Download the script:
```script
go get github.com/h-marvin/GoGit
```

Navigate to the folder folder where you want the build Go artefact to be stored (e.g. ~/Scripts/). Then start the build process:
```script
go build $GOPATH/src/github.com/h-marvin/GoGit/git-update.go
```

## Usage
This is the easy part. Just go ahead and call the script with the options you like. Example:
```script
~/Scripts/GoGit -recursive=true
```

### Available flags
To customize the script at execution for your needs, there are up to four flags that can be added.

`-path` allows to specify the root folder from where the script will start looking for git repos (defaults to user home). Multiple root locations can be added by separating them with a colon.

`-filter` allows filtering for a certain value to occur in the .git/config (e.g. enterprise git address)

`-recursive` specifies whether or not only direct children of _-path_ will be checked or the entire subtree

`-fetch` if set to _true_ a _fetch --prune_ will be performed instead of a _pull_
