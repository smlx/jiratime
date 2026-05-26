package client

import (
	"context"
	"fmt"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"golang.org/x/exp/slog"
)

// getAllIssues returns all the issues by automatically paging through results.
func getAllIssues(ctx context.Context, c *jira.Client,
	search string) ([]jira.Issue, error) {
	var issues []jira.Issue
	var nextPageToken string
	for {
		opt := &jira.SearchOptionsV2{
			MaxResults:    1000, // max 1000
			Fields:        []string{"id", "key"},
			NextPageToken: nextPageToken,
		}
		chunk, resp, err := c.Issue.SearchV2JQL(ctx, search, opt)
		if err != nil {
			return nil, fmt.Errorf("couldn't search: %v", err)
		}
		issues = append(issues, chunk...)
		// nextPageToken is null on the initial request, and on the last page
		if len(resp.NextPageToken) == 0 {
			return issues, nil
		}
		nextPageToken = resp.NextPageToken
	}
}

type worklogOpts struct {
	jira.SearchOptions
	StartedAfter int64 `url:"startedAfter,omitempty"`
}

// getWorklogRecords returns all worklogs where the author is the given user
// on the given issue.
func getWorklogRecords(ctx context.Context, c *jira.Client,
	issueID, authorEmail string, since time.Time) ([]jira.WorklogRecord, error) {
	last := 0
	var wlrs []jira.WorklogRecord
	for {
		opt := worklogOpts{
			SearchOptions: jira.SearchOptions{
				MaxResults: 5000, // max 5000
				StartAt:    last,
			},
			StartedAfter: since.UnixMilli(),
		}
		worklog, resp, err := c.Issue.GetWorklogs(ctx, issueID,
			jira.WithQueryOptions(&opt))
		if err != nil {
			return nil, fmt.Errorf("couldn't search: %v", err)
		}
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("bad response %d: %v", resp.StatusCode, resp.Status)
		}
		// filter the worklog records by author
		for _, wlr := range worklog.Worklogs {
			if wlr.Author.EmailAddress == authorEmail {
				wlrs = append(wlrs, wlr)
			}
		}
		last = resp.StartAt + len(worklog.Worklogs)
		// return if we have paged through all records otherwise get the next page
		// NOTE: the jira library is buggy and returns resp.Total == 0 here, so
		// check for "last page" indirectly.
		if len(worklog.Worklogs) < opt.MaxResults {
			return wlrs, nil
		}
	}
}

// Worklogs returns the worklogs since the given time.
func Worklogs(ctx context.Context, log *slog.Logger, c *jira.Client, userEmail string,
	since time.Time) (map[string][]jira.WorklogRecord, error) {
	// get all the issues with a worklog by the author
	issues, err := getAllIssues(ctx, c,
		fmt.Sprintf(`worklogAuthor = currentUser() AND worklogDate >= "%s"`,
			since.Format("2006-01-02")))
	if err != nil {
		return nil, fmt.Errorf("couldn't get issues: %v", err)
	}
	log.InfoCtx(ctx, "found issues", slog.Int("issueCount", len(issues)))
	// iterate through the issues getting all the associated worklogs
	worklogs := map[string][]jira.WorklogRecord{}
	for _, issue := range issues {
		wlrs, err := getWorklogRecords(ctx, c, issue.Key, userEmail, since)
		if err != nil {
			return nil, fmt.Errorf("couldn't get worklogs: %v", err)
		}
		log.InfoCtx(ctx, "found worklog records",
			slog.String("issue", issue.Key),
			slog.Int("worklogRecordCount", len(wlrs)))
		worklogs[issue.Key] = wlrs
	}
	return worklogs, nil
}
