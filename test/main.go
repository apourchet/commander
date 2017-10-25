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
