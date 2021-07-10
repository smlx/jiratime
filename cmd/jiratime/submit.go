package main

import (
	"fmt"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/smlx/jiratime/internal/client"
	"github.com/smlx/jiratime/internal/config"
	"github.com/smlx/jiratime/internal/parse"
)

// SubmitCmd represents the default `submit` command.
type SubmitCmd struct{}

// Run the Submit command.
func (cmd *SubmitCmd) Run() error {
	ctx, cancel := getContext(8 * time.Second)
	defer cancel()
	// read config file
	conf, err := config.Load(`./config.yml`)
	if err != nil {
		return fmt.Errorf("couldn't load config: %v", err)
	}
	spew.Dump(conf)
	// parse each line of input, generating a map of jira tickets
	// with associated Worklog entries
	worklogs, err := parse.Input(os.Stdin, conf)
	if err != nil {
		return fmt.Errorf("couldn't parse worklogs: %v", err)
	}
	spew.Dump(worklogs)
	// push the worklogs into jira
	if err = client.UploadWorklogs(ctx, worklogs, 0); err != nil {
		return fmt.Errorf("couldn't upload worklogs: %v", err)
	}
	return nil
}
