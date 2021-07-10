package main

import (
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/smlx/jiratime/internal/client"
	"github.com/smlx/jiratime/internal/config"
	"github.com/smlx/jiratime/internal/parse"
)

func main() {
	// read config file
	conf, err := config.Load(`./config.yml`)
	if err != nil {
		log.Fatalf("couldn't load config: %v", err)
	}
	spew.Dump(conf)
	// parse each line of input, generating a map of jira tickets
	// with associated Worklog entries
	worklogs, err := parse.Input(os.Stdin, conf)
	if err != nil {
		log.Fatalf("couldn't parse worklogs: %v", err)
	}
	spew.Dump(worklogs)
	// push the worklogs into jira
	if err = client.UploadWorklogs(worklogs, 0); err != nil {
		log.Fatalf("couldn't upload worklogs: %v", err)
	}
}
