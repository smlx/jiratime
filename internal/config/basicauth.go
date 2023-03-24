package config

import (
	"fmt"
	"os"

	"github.com/adrg/xdg"
	"sigs.k8s.io/yaml"
)

const basicAuthPathSuffix = "jiratime/basicauth.yml"

// BasicAuth represents the structure of the auth.yml
type BasicAuth struct {
	User   string `json:"user"`
	APIKey string `json:"apiKey"`
}

// ReadBasicAuth the config file.
func ReadBasicAuth() (*BasicAuth, error) {
	path, err := xdg.ConfigFile(basicAuthPathSuffix)
	if err != nil {
		return nil, fmt.Errorf("couldn't get path to basicAuth config file: %v", err)
	}
	y, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't read basicAuth config file: %v", err)
	}
	var a BasicAuth
	if err = yaml.Unmarshal(y, &a); err != nil {
		return nil, fmt.Errorf("couldn't unmarshal basicAuth config: %v", err)
	}
	return &a, nil
}

// WriteBasicAuth persists the given Config to the given path.
func WriteBasicAuth(a *BasicAuth) error {
	path, err := xdg.ConfigFile(basicAuthPathSuffix)
	if err != nil {
		return fmt.Errorf("couldn't get path to basicAuth config file: %v", err)
	}
	confBytes, err := yaml.Marshal(a)
	if err != nil {
		return fmt.Errorf("couldn't marshal basicAuth config: %v", err)
	}
	if err = os.WriteFile(path, confBytes, 0600); err != nil {
		return fmt.Errorf("couldn't write basicAuth config file: %v", err)
	}
	return nil
}
