package client

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	jira "github.com/andygrunwald/go-jira"
	"github.com/smlx/jiratime/internal/config"
	"github.com/smlx/jiratime/internal/parse"
	"golang.org/x/oauth2"
)

const requestRetries = 4

// authenticatedRoundTripper implements the http.RoundTripper interface
type authenticatedRoundTripper struct {
	username string
	password string
}

// RoundTrip sets the basic authentication header and then handles the request
// using the http.DefaultTransport.
func (art *authenticatedRoundTripper) RoundTrip(
	req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(art.username, art.password)
	return http.DefaultTransport.RoundTrip(req)
}

func newBasicAuthHTTPClient() (*http.Client, error) {
	basic, err := config.ReadBasicAuth()
	if err != nil {
		return nil, fmt.Errorf("couldn't read basic auth: %v", err)
	}
	// construct http.Client with automatic basic auth
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &authenticatedRoundTripper{
			username: basic.User,
			password: basic.APIKey,
		},
	}, nil
}

func newOAuth2HTTPClient(ctx context.Context) (*http.Client, oauth2.TokenSource, *config.OAuth2, error) {
	// load the auth config to get the oauth2 token
	auth, err := config.ReadAuth()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("couldn't load auth config: %v", err)
	}
	// sanity check that there is an access_token and refresh_token
	if auth == nil {
		return nil, nil, nil, fmt.Errorf("couldn't find oauth2 configuration")
	}
	if auth.Token.AccessToken == "" || auth.Token.RefreshToken == "" {
		return nil, nil, nil, fmt.Errorf("missing access_token or refresh_token." +
			" Please run `authorize` to refresh tokens")
	}
	// create an http client using the oauth2 token. this will auto-refresh the
	// token as required.
	oauth2Conf := GetOAuth2Config(auth)
	tokenSource := oauth2Conf.TokenSource(ctx, auth.Token)
	httpClient := oauth2.NewClient(ctx, tokenSource)
	return httpClient, tokenSource, auth, nil
}

// UploadWorklogs uploads the given worklogs to Jira, with the
// given day offset (e.g. -1 == yesterday).
func UploadWorklogs(ctx context.Context, jiraURL string,
	issueWorklogs map[string][]parse.Worklog, dayOffset int, dryRun bool,
	basicAuth bool) error {
	var httpClient *http.Client
	var err error
	var tokenSource oauth2.TokenSource
	var auth *config.OAuth2
	if basicAuth {
		httpClient, err = newBasicAuthHTTPClient()
		if err != nil {
			return fmt.Errorf("couldn't construct basic auth HTTP client: %v", err)
		}
	} else {
		httpClient, tokenSource, auth, err = newOAuth2HTTPClient(ctx)
		if err != nil {
			return fmt.Errorf("couldn't construct OAuth2 HTTP client: %v", err)
		}
	}
	// wrap this http client in a jira client via jira.NewClient
	c, err := jira.NewClient(httpClient, jiraURL)
	if err != nil {
		return fmt.Errorf("couldn't get new Jira client: %v", err)
	}
	// check that all the issues in worklogs exist
	var success bool
	for issue := range issueWorklogs {
		success = false
		var err error
		var response *jira.Response
		for i := 0; i < requestRetries; i++ {
			_, response, err = c.Issue.GetWithContext(ctx, issue, nil)
			if err == nil {
				success = true
				break // successful get
			}
			if response != nil && response.StatusCode == http.StatusUnauthorized {
				continue
			}
			return fmt.Errorf("couldn't get Jira issue %s: %v", issue, err)
		}
		if !success {
			return fmt.Errorf("couldn't get Jira issue %s after %d retries: %v",
				issue, requestRetries, err)
		}
	}
	if !basicAuth {
		// persist the token, as Atlassian rotates refresh tokens
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
	// exit early in dry-run mode
	if dryRun {
		log.Println("dry-run mode: not submitting any work logs")
		return nil
	}
	// add the worklogs to the issues
	for issue, worklogs := range issueWorklogs {
		for _, worklog := range worklogs {
			started := jira.Time(
				worklog.Started.Add(time.Hour * 24 * time.Duration(dayOffset)))
			wr := jira.WorklogRecord{
				Comment:          worklog.Comment,
				TimeSpentSeconds: int(worklog.Duration.Seconds()),
				Started:          &started,
			}
		retryAddWorklog:
			for i := 0; i < requestRetries; i++ {
				_, response, err := c.Issue.AddWorklogRecordWithContext(ctx, issue, &wr)
				if err == nil {
					break retryAddWorklog
				}
				if response.StatusCode == http.StatusNotFound {
					continue
				}
				return fmt.Errorf("couldn't add worklog record to issue %s: %v", issue, err)
			}
		}
	}
	return nil
}
