package client

import (
	"context"
	"fmt"
	"log"
	"time"

	jira "github.com/andygrunwald/go-jira"
	"github.com/smlx/jiratime/internal/config"
	"github.com/smlx/jiratime/internal/parse"
	"golang.org/x/oauth2"
)

// UploadWorklogs uploads the given worklogs to Jira, with the
// given day offset (e.g. -1 == yesterday).
func UploadWorklogs(ctx context.Context, jiraURL string,
	issueWorklogs map[string][]parse.Worklog, dayOffset int, dryRun bool) error {
	// load the auth config to get the oauth2 token
	auth, err := config.ReadAuth()
	if err != nil {
		return fmt.Errorf("couldn't load auth config: %v", err)
	}
	// sanity check that there is an access_token and refresh_token
	if auth == nil {
		return fmt.Errorf("couldn't find oauth2 configuration")
	}
	if auth.Token.AccessToken == "" || auth.Token.RefreshToken == "" {
		return fmt.Errorf("missing access_token or refresh_token. Please run `authorize` to refresh tokens")
	}
	// create an http client using the oauth2 token. this will auto-refresh the
	// token as required.
	oauth2Conf := GetOAuth2Config(auth)
	tokenSource := oauth2Conf.TokenSource(ctx, auth.Token)
	httpClient := oauth2.NewClient(ctx, tokenSource)
	// wrap this http client in a jira client via jira.NewClient
	c, err := jira.NewClient(httpClient, jiraURL)
	if err != nil {
		return fmt.Errorf("couldn't get new Jira client: %v", err)
	}
	// check that all the issues in worklogs exist
	for issue := range issueWorklogs {
		_, _, err := c.Issue.GetTransitionsWithContext(ctx, issue)
		if err != nil {
			return fmt.Errorf("couldn't get Jira issue %s: %v", issue, err)
		}
	}
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
			_, _, err := c.Issue.AddWorklogRecordWithContext(ctx, issue, &wr)
			if err != nil {
				return fmt.Errorf("couldn't add worklog record to issue %s: %v", issue, err)
			}
		}
	}
	return nil
}
