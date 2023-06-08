// Package process performs post-processing on parsed worklogs.
package process

import (
	"time"

	"github.com/smlx/jiratime/internal/config"
	"github.com/smlx/jiratime/internal/parse"
)

// getRoundTime returns the duration which needs to be added to bring the slice
// of Worklogs up to a multiple of 15 minutes.
func getRoundTime(worklogs []parse.Worklog) time.Duration {
	// sum the worklogs durations
	var total time.Duration
	for _, worklog := range worklogs {
		total += worklog.Duration
	}
	// mod 15 minutes
	mod15 := total % (15 * time.Minute)
	// if not zero, subtract from 15 minutes
	if mod15 > 0 {
		return (15 * time.Minute) - mod15
	}
	return 0
}

// RoundWorklogs takes a map of parsed worklogs and returns an updated map with
// the total worklogs of matching issues rounded up to the next 15 minutes.
//
// This is done by adding a worklog entry for the issue to round out the total.
func RoundWorklogs(worklogs map[string][]parse.Worklog,
	roundIssues []config.Regexp) {
	for issueKey := range worklogs {
		for _, roundIssue := range roundIssues {
			if roundIssue.MatchString(issueKey) {
				roundTime := getRoundTime(worklogs[issueKey])
				if roundTime > 0 {
					// add a rounding issue to the issue worklogs
					now := time.Now()
					worklogs[issueKey] = append(worklogs[issueKey],
						parse.Worklog{
							Started: time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0,
								0, now.Location()),
							Duration: roundTime,
							Comment:  "round to 15 minutes",
						})
				}
				break // go to the next issueKey
			}
		}
	}
}
