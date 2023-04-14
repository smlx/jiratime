package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	jira "github.com/andygrunwald/go-jira"
	"golang.org/x/exp/slog"
)

// getAllIssues returns all the issues by automatically paging through results.
func getAllIssues(ctx context.Context, c *jira.Client,
	search string) ([]jira.Issue, error) {
	last := 0
	var issues []jira.Issue
	for {
		opt := &jira.SearchOptions{
			MaxResults: 1000, // max 1000
			StartAt:    last,
			Fields:     []string{"id", "key"},
		}
		chunk, resp, err := c.Issue.SearchWithContext(ctx, search, opt)
		if err != nil {
			return nil, fmt.Errorf("couldn't search: %v", err)
		}
		total := resp.Total
		issues = append(issues, chunk...)
		last = resp.StartAt + len(chunk)
		if last >= total {
			return issues, nil
		}
	}
}

type worklogOpts struct {
	jira.SearchOptions
	StartedAfter int64 `url:"startedAfter,omitempty"`
}

// getWorklogRecords returns all worklogs where the author is the given user
// on the given issue.
func getWorklogRecords(ctx context.Context, c *jira.Client,
	issueID, authorName string, since time.Time) ([]jira.WorklogRecord, error) {
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
		worklog, resp, err := c.Issue.GetWorklogsWithContext(ctx, issueID,
			jira.WithQueryOptions(&opt))
		if err != nil {
			return nil, fmt.Errorf("couldn't search: %v", err)
		}
		total := resp.Total
		// filter the worklog records by author
		for _, wlr := range worklog.Worklogs {
			if wlr.Author.DisplayName == authorName {
				wlrs = append(wlrs, wlr)
			}
		}
		last = resp.StartAt + len(worklog.Worklogs)
		// return if we have paged through all records
		if last >= total {
			return wlrs, nil
		}
	}
}

// Worklogs returns the worklogs since the given time.
func Worklogs(ctx context.Context, log *slog.Logger, jiraURL string,
	since time.Time, authorName string,
	basicAuth bool) (map[string][]jira.WorklogRecord, error) {
	var httpClient *http.Client
	var err error
	httpClient, err = newBasicAuthHTTPClient()
	if err != nil {
		return nil, fmt.Errorf(
			"couldn't construct basic auth HTTP client (required for worklogs): %v", err)
	}
	// wrap this http client in a jira client via jira.NewClient
	c, err := jira.NewClient(httpClient, jiraURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't get new Jira client: %v", err)
	}
	// get all the issues with a worklog by the author
	issues, err := getAllIssues(ctx, c,
		fmt.Sprintf(`worklogAuthor = "%s" AND worklogDate >= "%s"`,
			authorName, since.Format("2006-01-02")))
	if err != nil {
		return nil, fmt.Errorf("couldn't get issues: %v", err)
	}
	log.InfoCtx(ctx, "found issues", slog.Int("issueCount", len(issues)))
	// iterate through the issues getting all the associated worklogs
	worklogs := map[string][]jira.WorklogRecord{}
	for _, issue := range issues {
		wlrs, err := getWorklogRecords(ctx, c, issue.Key, authorName, since)
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
