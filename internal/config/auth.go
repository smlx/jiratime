package config

import (
	"fmt"
	"os"

	"github.com/adrg/xdg"
	"golang.org/x/oauth2"
	"sigs.k8s.io/yaml"
)

const authPathSuffix = "jiratime/auth.yml"

// Auth represents the structure of the auth.yml
type Auth struct {
	OAuth2 *OAuth2 `json:"oauth2"`
}

// OAuth2 is a config entry containing oauth2 secrets
type OAuth2 struct {
	ClientID string        `json:"clientID,omitempty"`
	Secret   string        `json:"secret,omitempty"`
	Token    *oauth2.Token `json:"token"`
}

// ReadAuth the config file.
func ReadAuth() (*OAuth2, error) {
	path, err := xdg.ConfigFile(authPathSuffix)
	if err != nil {
		return nil, fmt.Errorf("couldn't get path to auth config file: %v", err)
	}
	y, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't read config file: %v", err)
	}
	var a Auth
	if err = yaml.Unmarshal(y, &a); err != nil {
		return nil, fmt.Errorf("couldn't unmarshal config: %v", err)
	}
	return a.OAuth2, nil
}

// WriteAuth persists the given Config to the given path.
func WriteAuth(o *OAuth2) error {
	path, err := xdg.ConfigFile(authPathSuffix)
	if err != nil {
		return fmt.Errorf("couldn't get path to auth config file: %v", err)
	}
	confBytes, err := yaml.Marshal(&Auth{OAuth2: o})
	if err != nil {
		return fmt.Errorf("couldn't marshal config: %v", err)
	}
	if err = os.WriteFile(path, confBytes, 0600); err != nil {
		return fmt.Errorf("couldn't write config file: %v", err)
	}
	return nil
}
