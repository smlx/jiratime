package config

import (
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"sigs.k8s.io/yaml"
)

// Issue represents the list of known JIRA issues.
type Issue struct {
	ID      string   `json:"id"`
	Regexes []Regexp `json:"regexes"`
}

// OAuth2 is a config entry containing oauth2 secrets
type OAuth2 struct {
	ClientID string        `json:"clientID"`
	Secret   string        `json:"secret"`
	Token    *oauth2.Token `json:"token"`
}

// Config represents the structure of the config file.
type Config struct {
	Issues []Issue `json:"issues"`
	OAuth2 *OAuth2 `json:"oauth2"`
}

// Load the config file.
func Load(path string) (*Config, error) {
	y, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't read config file: %v", err)
	}
	var c Config
	if err = yaml.Unmarshal(y, &c); err != nil {
		return nil, fmt.Errorf("couldn't unmarshal config file: %v", err)
	}
	return &c, nil
}
