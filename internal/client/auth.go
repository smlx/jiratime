// Package client implements a JIRA REST API client.
package client

import (
	"github.com/smlx/jiratime/internal/config"
	"golang.org/x/oauth2"
)

// GetOAuth2Config gets an OAuth2 Config object configured for Atlassian Jira
// Cloud.
func GetOAuth2Config(auth *config.OAuth2) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     auth.ClientID,
		ClientSecret: auth.Secret,
		// offline_access requests a refresh token
		Scopes: []string{
			"offline_access",
			"read:jira-work",
			"write:jira-work",
		},
		Endpoint: oauth2.Endpoint{
			TokenURL: "https://auth.atlassian.com/oauth/token",
			AuthURL:  "https://auth.atlassian.com/authorize",
		},
		RedirectURL: "http://localhost:8080/oauth/redirect",
	}
}
