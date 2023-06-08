package process

import (
	"regexp"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/smlx/jiratime/internal/config"
	"github.com/smlx/jiratime/internal/parse"
)

func TestGetRoundTime(t *testing.T) {
	var testCases = map[string]struct {
		input  []parse.Worklog
		expect time.Duration
	}{
		"nil slice": {
			input:  nil,
			expect: 0,
		},
		"nothing logged": {
			input:  []parse.Worklog{},
			expect: 0,
		},
		"less than 15 min logged in a single entry": {
			input: []parse.Worklog{
				{Duration: 5 * time.Minute},
			},
			expect: 10 * time.Minute,
		},
		"less than 15 min logged in multiple entries": {
			input: []parse.Worklog{
				{Duration: 5 * time.Minute},
				{Duration: 5 * time.Minute},
			},
			expect: 5 * time.Minute,
		},
		"multiple of 15 min logged in single entry": {
			input: []parse.Worklog{
				{Duration: 15 * time.Minute},
			},
			expect: 0,
		},
		"multiple of 15 min logged in multiple entries 1": {
			input: []parse.Worklog{
				{Duration: 5 * time.Minute},
				{Duration: 10 * time.Minute},
			},
			expect: 0,
		},
		"multiple of 15 min logged in multiple entries 2": {
			input: []parse.Worklog{
				{Duration: 5 * time.Minute},
				{Duration: 5 * time.Minute},
				{Duration: 5 * time.Minute},
			},
			expect: 0,
		},
		"more than 15 min logged in single entry": {
			input: []parse.Worklog{
				{Duration: 40 * time.Minute},
			},
			expect: 5 * time.Minute,
		},
		"more than 15 min logged in multiple entries": {
			input: []parse.Worklog{
				{Duration: 5 * time.Minute},
				{Duration: 5 * time.Minute},
				{Duration: 75 * time.Minute},
			},
			expect: 5 * time.Minute,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			assert.Equal(tt, tc.expect, getRoundTime(tc.input), "getRoundTime")
		})
	}
}

func TestRoundWorklogs(t *testing.T) {
	now := time.Now()
	type roundWorklogsInput struct {
		worklogs    map[string][]parse.Worklog
		roundIssues []config.Regexp
	}
	var testCases = map[string]struct {
		input  roundWorklogsInput
		expect map[string][]parse.Worklog
	}{
		"nil worklogs": {
			input: roundWorklogsInput{
				worklogs: nil,
				roundIssues: []config.Regexp{
					{Regexp: *regexp.MustCompile("^FOO-")},
				},
			}, expect: nil,
		},
		"empty worklogs": {
			input: roundWorklogsInput{
				worklogs: map[string][]parse.Worklog{},
				roundIssues: []config.Regexp{
					{Regexp: *regexp.MustCompile("^FOO-")},
				},
			}, expect: map[string][]parse.Worklog{},
		},
		"nil roundIssues": {
			input: roundWorklogsInput{
				worklogs:    map[string][]parse.Worklog{},
				roundIssues: nil,
			}, expect: map[string][]parse.Worklog{},
		},
		"empty roundIssues": {
			input: roundWorklogsInput{
				worklogs:    map[string][]parse.Worklog{},
				roundIssues: []config.Regexp{},
			}, expect: map[string][]parse.Worklog{},
		},
		"single issue, match, no rounding": {
			input: roundWorklogsInput{
				worklogs: map[string][]parse.Worklog{
					"FOO-12": {
						{Duration: 15 * time.Minute},
					},
				},
				roundIssues: []config.Regexp{
					{Regexp: *regexp.MustCompile("^FOO-")},
				},
			},
			expect: map[string][]parse.Worklog{
				"FOO-12": {
					{Duration: 15 * time.Minute},
				},
			},
		},
		"single issue, match, rounding": {
			input: roundWorklogsInput{
				worklogs: map[string][]parse.Worklog{
					"FOO-12": {
						{Duration: 20 * time.Minute},
					},
				},
				roundIssues: []config.Regexp{
					{Regexp: *regexp.MustCompile("^FOO-")},
				},
			},
			expect: map[string][]parse.Worklog{
				"FOO-12": {
					{Duration: 20 * time.Minute},
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0,
							0, now.Location()),
						Duration: 10 * time.Minute,
						Comment:  "round to 15 minutes",
					},
				},
			},
		},
		"single issue, no match": {
			input: roundWorklogsInput{
				worklogs: map[string][]parse.Worklog{
					"ABC-4": {
						{Duration: 20 * time.Minute},
					},
				},
				roundIssues: []config.Regexp{
					{Regexp: *regexp.MustCompile("^FOO-")},
				},
			},
			expect: map[string][]parse.Worklog{
				"ABC-4": {
					{Duration: 20 * time.Minute},
				},
			},
		},
		"multiple issues, single match, rounding": {
			input: roundWorklogsInput{
				worklogs: map[string][]parse.Worklog{
					"ABC-4": {
						{Duration: 20 * time.Minute},
					},
					"FOO-66": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
					},
				},
				roundIssues: []config.Regexp{
					{Regexp: *regexp.MustCompile("^FOO-")},
				},
			},
			expect: map[string][]parse.Worklog{
				"ABC-4": {
					{Duration: 20 * time.Minute},
				},
				"FOO-66": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0,
							0, now.Location()),
						Duration: 5 * time.Minute,
						Comment:  "round to 15 minutes",
					},
				},
			},
		},
		"multiple issues, single match, no rounding": {
			input: roundWorklogsInput{
				worklogs: map[string][]parse.Worklog{
					"ABC-4": {
						{Duration: 20 * time.Minute},
					},
					"FOO-86": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
					},
				},
				roundIssues: []config.Regexp{
					{Regexp: *regexp.MustCompile("^FOO-")},
				},
			},
			expect: map[string][]parse.Worklog{
				"ABC-4": {
					{Duration: 20 * time.Minute},
				},
				"FOO-86": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
				},
			},
		},
		"multiple issues, multiple match, some rounding": {
			input: roundWorklogsInput{
				worklogs: map[string][]parse.Worklog{
					"ABC-4": {
						{Duration: 20 * time.Minute},
					},
					"FOO-86": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
					},
					"FOO-96": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 5 * time.Minute},
					},
					"FOO-92": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
					},
					"FOO-56": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 40 * time.Minute},
					},
				},
				roundIssues: []config.Regexp{
					{Regexp: *regexp.MustCompile("^FOO-")},
				},
			},
			expect: map[string][]parse.Worklog{
				"ABC-4": {
					{Duration: 20 * time.Minute},
				},
				"FOO-86": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
				},
				"FOO-96": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 5 * time.Minute},
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0,
							0, now.Location()),
						Duration: 10 * time.Minute,
						Comment:  "round to 15 minutes",
					},
				},
				"FOO-92": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
				},
				"FOO-56": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 40 * time.Minute},
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0,
							0, now.Location()),
						Duration: 5 * time.Minute,
						Comment:  "round to 15 minutes",
					},
				},
			},
		},
		"multiple issues, multiple match, no rounding": {
			input: roundWorklogsInput{
				worklogs: map[string][]parse.Worklog{
					"ABC-4": {
						{Duration: 20 * time.Minute},
					},
					"FOO-86": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
					},
					"FOO-96": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
					},
					"FOO-92": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 15 * time.Minute},
					},
					"FOO-56": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 45 * time.Minute},
					},
				},
				roundIssues: []config.Regexp{
					{Regexp: *regexp.MustCompile("^FOO-")},
				},
			},
			expect: map[string][]parse.Worklog{
				"ABC-4": {
					{Duration: 20 * time.Minute},
				},
				"FOO-86": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
				},
				"FOO-96": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
				},
				"FOO-92": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 15 * time.Minute},
				},
				"FOO-56": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 45 * time.Minute},
				},
			},
		},
		"multiple issues, all match, some rounding": {
			input: roundWorklogsInput{
				worklogs: map[string][]parse.Worklog{
					"FOO-86": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
					},
					"FOO-96": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 10 * time.Minute},
					},
					"FOO-92": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 10 * time.Minute},
					},
					"FOO-56": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 45 * time.Minute},
					},
				},
				roundIssues: []config.Regexp{
					{Regexp: *regexp.MustCompile("^FOO-")},
				},
			},
			expect: map[string][]parse.Worklog{
				"FOO-86": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
				},
				"FOO-96": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 10 * time.Minute},
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0,
							0, now.Location()),
						Duration: 10 * time.Minute,
						Comment:  "round to 15 minutes",
					},
				},
				"FOO-92": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 10 * time.Minute},
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0,
							0, now.Location()),
						Duration: 5 * time.Minute,
						Comment:  "round to 15 minutes",
					},
				},
				"FOO-56": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 45 * time.Minute},
				},
			},
		},
		"multiple issues, all match, all rounding": {
			input: roundWorklogsInput{
				worklogs: map[string][]parse.Worklog{
					"FOO-86": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 15 * time.Minute},
					},
					"FOO-96": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 10 * time.Minute},
					},
					"FOO-92": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 10 * time.Minute},
					},
					"FOO-56": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 40 * time.Minute},
					},
				},
				roundIssues: []config.Regexp{
					{Regexp: *regexp.MustCompile("^FOO-")},
				},
			},
			expect: map[string][]parse.Worklog{
				"FOO-86": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 15 * time.Minute},
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0,
							0, now.Location()),
						Duration: 5 * time.Minute,
						Comment:  "round to 15 minutes",
					},
				},
				"FOO-96": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 10 * time.Minute},
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0,
							0, now.Location()),
						Duration: 10 * time.Minute,
						Comment:  "round to 15 minutes",
					},
				},
				"FOO-92": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 10 * time.Minute},
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0,
							0, now.Location()),
						Duration: 5 * time.Minute,
						Comment:  "round to 15 minutes",
					},
				},
				"FOO-56": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 40 * time.Minute},
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0,
							0, now.Location()),
						Duration: 5 * time.Minute,
						Comment:  "round to 15 minutes",
					},
				},
			},
		},
		"multiple issues, all match, no rounding": {
			input: roundWorklogsInput{
				worklogs: map[string][]parse.Worklog{
					"FOO-86": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 15 * time.Minute},
						{Duration: 5 * time.Minute},
					},
					"FOO-96": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 5 * time.Minute},
					},
					"FOO-92": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
					},
					"FOO-56": {
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 20 * time.Minute},
						{Duration: 30 * time.Minute},
					},
				},
				roundIssues: []config.Regexp{
					{Regexp: *regexp.MustCompile("^FOO-")},
				},
			},
			expect: map[string][]parse.Worklog{
				"FOO-86": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 15 * time.Minute},
					{Duration: 5 * time.Minute},
				},
				"FOO-96": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 5 * time.Minute},
				},
				"FOO-92": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
				},
				"FOO-56": {
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 20 * time.Minute},
					{Duration: 30 * time.Minute},
				},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			RoundWorklogs(tc.input.worklogs, tc.input.roundIssues)
			assert.Equal(tt, tc.expect, tc.input.worklogs, "RoundWorklogs")
		})
	}
}
