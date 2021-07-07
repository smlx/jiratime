package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"sigs.k8s.io/yaml"
)

// Issue represents the list of known JIRA issues.
type Issue struct {
	ID      string   `json:"id"`
	Regexes []Regexp `json:"regexes"`
}

// Config represents the structure of the config file.
type Config struct {
	Issues []Issue `json:"issues"`
}

// Worklog represents an individual work log entry on a ticket.
type Worklog struct {
	Duration time.Duration
	Comment string // optional
}

func loadConfig() (*Config, error) {
	y, err := os.ReadFile(`./config.yml`)
	if err != nil {
		return nil, fmt.Errorf("couldn't read config file: %v", err)
	}
	var c Config
	if err = yaml.Unmarshal(y, &c); err != nil {
		return nil, fmt.Errorf("couldn't unmarshal config file: %v", err)
	}
	return &c, nil
}

func main() {
	// read config file
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("couldn't load config: %v", err)
	}
	spew.Dump(config)

	// parse each line of input, generating a map of jira tickets with associated Worklog entries
	worklogs, err := ParseInput(os.Stdin, config)
	if err != nil {
		log.Fatalf("couldn't parse worklogs: %v", err)
	}
	spew.Dump(worklogs)

	// iterate over the time entries and comments, adding the time to jira
}
