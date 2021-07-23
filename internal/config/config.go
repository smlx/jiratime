package config

import (
	"fmt"
	"os"

	"github.com/adrg/xdg"
	"sigs.k8s.io/yaml"
)

const pathSuffix = "jiratime/config.yml"

// Issue represents the list of known JIRA issues.
type Issue struct {
	ID             string   `json:"id"`
	Regexes        []Regexp `json:"regexes"`
	DefaultComment string   `json:"defaultComment"`
}

// Config represents the structure of the config file.
type Config struct {
	JiraURL string   `json:"jiraURL"`
	Issues  []Issue  `json:"issues"`
	Ignore  []Regexp `json:"ignore"`
}

// Read the config file.
func Read() (*Config, error) {
	path, err := xdg.ConfigFile(pathSuffix)
	if err != nil {
		return nil, fmt.Errorf("couldn't get path to config file: %v", err)
	}
	y, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't read config file: %v", err)
	}
	var c Config
	if err = yaml.Unmarshal(y, &c); err != nil {
		return nil, fmt.Errorf("couldn't unmarshal config: %v", err)
	}
	return &c, nil
}

// Write persists the given Config to the given path.
func Write(c *Config) error {
	path, err := xdg.ConfigFile(pathSuffix)
	if err != nil {
		return fmt.Errorf("couldn't get path to config file: %v", err)
	}
	confBytes, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("couldn't marshal config: %v", err)
	}
	if err = os.WriteFile(path, confBytes, 0600); err != nil {
		return fmt.Errorf("couldn't write config file: %v", err)
	}
	return nil
}
