package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/smlx/jiratime/internal/client"
	"github.com/smlx/jiratime/internal/config"
	"golang.org/x/exp/slog"
)

// DumpWorklogsCmd represents the `dump-worklogs` command.
type DumpWorklogsCmd struct {
	Since     time.Time     `kong:"required,help='time from which the worklogs should be dumped'"`
	Timeout   time.Duration `kong:"default=1h,help='maximum duration allowed for the command to return'"`
	BasicAuth bool          `kong:"help='force basic auth instead of OAuth2'"`
}

// Run the DumpWorklogs command.
func (cmd *DumpWorklogsCmd) Run() error {
	// global timeout of 60 seconds
	ctx, cancel := getContext(cmd.Timeout)
	defer cancel()
	// read config file
	conf, err := config.Read()
	if err != nil {
		return fmt.Errorf("couldn't load config: %v", err)
	}
	level := slog.LevelVar{}
	level.Set(slog.LevelDebug)
	log := slog.New(
		slog.HandlerOptions{
			AddSource: true,
			Level:     &level,
		}.NewJSONHandler(os.Stderr))

	c, userEmail, persistToken, err := getJiraClient(ctx, conf.JiraURL, cmd.BasicAuth)
	if err != nil {
		return fmt.Errorf("couldn't get Jira client: %v", err)
	}

	// get the worklogs
	worklogs, err := client.Worklogs(ctx, log, c, userEmail, cmd.Since)
	if err != nil {
		return fmt.Errorf("couldn't dump worklogs: %v", err)
	}
	if err := persistToken(); err != nil {
		return fmt.Errorf("couldn't persist token: %v", err)
	}

	data, err := json.Marshal(worklogs)
	if err != nil {
		return fmt.Errorf("couldn't marshal worklogs: %v", err)
	}
	_, err = fmt.Println(string(data))
	return err
}
