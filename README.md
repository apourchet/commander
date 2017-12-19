# Commander
A CLI framework in Golang that leverages the `reflect` package to get rid of some annoying switch statements and argument parsing logic that most Golang CLI applications have to deal with.

# Example
The following example shows how one could implement bogus file manager. Examples of the usage would be the following:
```
$ # prints out the contents of that file
$ manager read filename
$ # prints out the contents of all the files (like cat)
$ manager read filename1 filename2
$ # deletes a file
$ manager rm filename
$ # mock a run of the rm command 
$ manager --dry-run=true rm filename
$ # uses a submodule of the cli to read from http
$ manager http read http://www.google.com
```
Below is what the code looks like. Notice that commander takes care of setting up the flags, and calling the right methods on your CLI objects depending on the command line arguments.
```go
// main.go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/apourchet/commander"
)

type Manager struct {
	// This DryRun field will be populated by command line flags. No more globals :)
	DryRun bool `commander:"flag=dry-run"`

	Http *HTTPCLI `commander:"subcommand=http"`
}

func (cli Manager) Read(file string, rest []string) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(content))
	for _, name := range rest {
		content, err := ioutil.ReadFile(name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(content))
	}
}

func (cli Manager) Rm(file string) {
	fmt.Println("Removing file", file)
	if cli.DryRun {
		return
	}
	os.Remove(file)
}

type HTTPCLI struct{}

func (cli HTTPCLI) Read(url string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(content))
}

func main() {
	cli := &Manager{
		Http: &HTTPCLI{},
	}
	cmd := commander.New()
	err := cmd.RunCLI(cli, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
}
```
If you want to try it out in this repository, the following commands work:
```
$ go run test/main.go read README.md
$ go run test/main.go read README.md README.md
$ go run test/main.go http read http://www.google.com
$ go run test/main.go -dry-run=true rm README.md
```
