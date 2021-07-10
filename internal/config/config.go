package config

import (
	"fmt"
	"os"

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

// Load the config file.
func Load(path string) (*Config, error) {
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
func Write(c *Config, path string) error {
	confBytes, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("couldn't marshal config: %v", err)
	}
	if err = os.WriteFile(path, confBytes, 0600); err != nil {
		return fmt.Errorf("couldn't write config file: %v", err)
	}
	return nil
}
