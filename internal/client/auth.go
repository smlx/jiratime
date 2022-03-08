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
			"read:avatar:jira",
			"read:field-configuration:jira",
			"read:group:jira",
			"read:issue-worklog.property:jira",
			"read:issue-worklog:jira",
			"read:issue.transition:jira",
			"read:project-role:jira",
			"read:status:jira",
			"read:user:jira",
			"write:issue-worklog.property:jira",
			"write:issue-worklog:jira",
			"write:issue.time-tracking:jira",
		},
		Endpoint: oauth2.Endpoint{
			TokenURL: "https://auth.atlassian.com/oauth/token",
			AuthURL:  "https://auth.atlassian.com/authorize",
		},
		RedirectURL: "http://localhost:8080/oauth/redirect",
	}
}
