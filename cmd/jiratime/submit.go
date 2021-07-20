package main

import (
	"fmt"
	"os"
	"time"

	"github.com/smlx/jiratime/internal/client"
	"github.com/smlx/jiratime/internal/config"
	"github.com/smlx/jiratime/internal/parse"
)

// SubmitCmd represents the default `submit` command.
type SubmitCmd struct {
	DayOffset int  `kong:"short='d',help='submit time for a day at some offset to today'"`
	DryRun    bool `kong:"help='read-only mode; do not actually make any changes in JIRA'"`
}

// Run the Submit command.
func (cmd *SubmitCmd) Run() error {
	// global timeout of 60 seconds
	ctx, cancel := getContext(60 * time.Second)
	defer cancel()
	// read config file
	conf, err := config.Read()
	if err != nil {
		return fmt.Errorf("couldn't load config: %v", err)
	}
	// parse each line of input, generating a map of jira tickets
	// with associated Worklog entries
	worklogs, err := parse.Input(os.Stdin, conf)
	if err != nil {
		return fmt.Errorf("couldn't parse worklogs: %v", err)
	}
	// push the worklogs into jira
	if err = client.UploadWorklogs(ctx, worklogs, cmd.DayOffset, cmd.DryRun); err != nil {
		return fmt.Errorf("couldn't upload worklogs: %v", err)
	}
	return nil
}
