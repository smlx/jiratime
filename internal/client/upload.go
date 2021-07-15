package client

import (
	"context"
	"fmt"
	"time"

	jira "github.com/andygrunwald/go-jira"
	"github.com/smlx/jiratime/internal/config"
	"github.com/smlx/jiratime/internal/parse"
)

// UploadWorklogs uploads the given worklogs to JIRA, with the
// given day offset (e.g. -1 == yesterday).
func UploadWorklogs(ctx context.Context,
	issueWorklogs map[string][]parse.Worklog, dayOffset int) error {
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
		return fmt.Errorf("Missing access_token or refresh_token. Please run `authorize` to refresh tokens")
	}
	// create an http client using the oauth2 token. this will auto-refresh the
	// token as required.
	oauth2Conf := GetOAuth2Config(auth)
	httpClient := oauth2Conf.Client(ctx, auth.Token)
	// wrap this http client in a jira client via jira.NewClient
	c, err := jira.NewClient(httpClient, "https://amazeeio.atlassian.net/")
	if err != nil {
		return fmt.Errorf("couldn't get new JIRA client: %v", err)
	}
	// check that all the issues in worklogs exist
	for issue := range issueWorklogs {
		_, _, err := c.Issue.GetWithContext(ctx, issue, nil)
		if err != nil {
			return fmt.Errorf("couldn't get JIRA issue %s: %v", issue, err)
		}
	}
	// calculate the started offset time in case we are using a day offset
	var started jira.Time = jira.Time(
		time.Now().Add(time.Hour * 24 * time.Duration(dayOffset)))
	// add the worklogs to the issues
	for issue, worklogs := range issueWorklogs {
		for _, worklog := range worklogs {
			wr := jira.WorklogRecord{
				Comment:          worklog.Comment,
				TimeSpentSeconds: int(worklog.Duration.Seconds()),
			}
			if dayOffset != 0 {
				wr.Started = &started
			}
			_, _, err := c.Issue.AddWorklogRecordWithContext(ctx, issue, &wr)
			if err != nil {
				return fmt.Errorf("couldn't add worklog record to issue %s: %v", issue, err)
			}
		}
	}
	// TODO: update the stored token in auth.yml? Not sure if it is even possible
	// to pull that back out of the httpClient...
	return nil
}
