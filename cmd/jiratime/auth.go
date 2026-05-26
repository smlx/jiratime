package main

import (
	"context"
	"fmt"
	"net/http"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/smlx/jiratime/internal/client"
	"github.com/smlx/jiratime/internal/config"
	"golang.org/x/oauth2"
)

// getJiraClient constructs an authenticated Jira client.
func getJiraClient(
	ctx context.Context,
	jiraURL string,
	basicAuthFlag bool,
) (*jira.Client, string, func() error, error) {
	useBasicAuth := basicAuthFlag || (config.HasBasicAuth() && !config.HasAuth())

	var httpClient *http.Client
	var err error
	var tokenSource oauth2.TokenSource
	var auth *config.OAuth2
	var userEmail string

	if useBasicAuth {
		httpClient, userEmail, err = client.NewBasicAuthHTTPClient()
		if err != nil {
			return nil, "", nil, fmt.Errorf("couldn't construct basic auth HTTP client: %v", err)
		}
	} else {
		httpClient, tokenSource, auth, err = client.NewOAuth2HTTPClient(ctx)
		if err != nil {
			return nil, "", nil, fmt.Errorf("couldn't construct OAuth2 HTTP client: %v", err)
		}

		jiraURL, err = client.OAuth2JiraURL(httpClient, jiraURL)
		if err != nil {
			return nil, "", nil, fmt.Errorf("couldn't construct OAuth2 Jira URL: %v", err)
		}
	}

	c, err := jira.NewClient(jiraURL, httpClient)
	if err != nil {
		return nil, "", nil, fmt.Errorf("couldn't get new Jira client: %v", err)
	}

	if !useBasicAuth {
		user, _, err := c.User.GetCurrentUser(ctx)
		if err != nil {
			return nil, "", nil, fmt.Errorf("couldn't get current user for OAuth2: %v", err)
		}
		userEmail = user.EmailAddress
	}

	persistToken := func() error {
		if !useBasicAuth {
			newTok, err := tokenSource.Token()
			if err != nil {
				return fmt.Errorf("couldn't get Token from oauth2.TokenSource: %v", err)
			}
			err = config.WriteAuth(&config.OAuth2{
				ClientID: auth.ClientID,
				Secret:   auth.Secret,
				Token:    newTok,
			})
			if err != nil {
				return fmt.Errorf("couldn't persist new token: %v", err)
			}
		}
		return nil
	}

	return c, userEmail, persistToken, nil
}
