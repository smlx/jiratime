package parse_test

import (
	"os"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/smlx/jiratime/internal/config"
	"github.com/smlx/jiratime/internal/parse"
)

type parseInput struct {
	dataFile string
	config   *config.Config
}

func wrapRegexes(regexes []string) []config.Regexp {
	var regexps []config.Regexp
	for _, s := range regexes {
		r := regexp.MustCompile(s)
		regexps = append(regexps, config.Regexp{Regexp: *r})
	}
	return regexps
}

func TestParseInput(t *testing.T) {
	now := time.Now()
	var testCases = map[string]struct {
		input  *parseInput
		expect map[string][]parse.Worklog
	}{
		"worklog0": {
			input: &parseInput{
				dataFile: "testdata/worklog0",
				config: &config.Config{
					Issues: []config.Issue{
						{
							ID: "PLATFORM-1",
							Regexes: wrapRegexes([]string{
								"^platform ops( .+)?$",
							}),
							DefaultComment: "platform ops",
						},
						{
							ID: "FOO-12",
							Regexes: wrapRegexes([]string{
								"^foo sync$",
							}),
							DefaultComment: "weekly catch-up with foo",
						},
						{
							ID: "FOO-3",
							Regexes: wrapRegexes([]string{
								"^fooCustomer devops( .+)?$",
							}),
							DefaultComment: "default fooCustomer work comment",
						},
						{
							ID: "INTERNAL-1",
							Regexes: wrapRegexes([]string{
								"^admin$",
							}),
							DefaultComment: "email / backlog grooming / slack",
						},
						{
							ID: "INTERNAL-2",
							Regexes: wrapRegexes([]string{
								"^standup$",
							}),
							DefaultComment: "standup",
						},
						{
							ID: "INTERNAL-3",
							Regexes: wrapRegexes([]string{
								"^pd$",
							}),
							DefaultComment: "primary on-call",
						},
						{
							ID: "BAR-1",
							Regexes: wrapRegexes([]string{
								"^bar sync$",
							}),
							DefaultComment: "bar customer weekly meeting",
						},
						{
							ID: "INTERNAL-4",
							Regexes: wrapRegexes([]string{
								"^platform sync$",
							}),
							DefaultComment: "platform sync",
						},
						{
							ID: "BAR-2",
							Regexes: wrapRegexes([]string{
								"^barCustomer infra( .+)?$",
							}),
						},
					},
				},
			},
			expect: map[string][]parse.Worklog{
				"PLATFORM-1": {
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 8, 40, 0,
							0, now.Location()),
						Duration: 20 * time.Minute,
						Comment:  "platform ops",
					},
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 16, 30, 0,
							0, now.Location()),
						Duration: 30 * time.Minute,
						Comment:  "example5 cluster melting down again",
					},
				},
				"FOO-12": {
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0,
							0, now.Location()),
						Duration: 30 * time.Minute,
						Comment:  "weekly catch-up with foo",
					},
				},
				"FOO-3": {
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 9, 30, 0,
							0, now.Location()),
						Duration: 70 * time.Minute,
						Comment:  "node scheduling issue",
					},
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 15, 30, 0,
							0, now.Location()),
						Duration: 15 * time.Minute,
						Comment:  "reply to MS",
					},
				},
				"INTERNAL-1": {
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 10, 40, 0,
							0, now.Location()),
						Duration: 80 * time.Minute,
						Comment:  "email / backlog grooming / slack",
					},
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 15, 45, 0,
							0, now.Location()),
						Duration: 5 * time.Minute,
						Comment:  "email / backlog grooming / slack",
					},
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 16, 0, 0,
							0, now.Location()),
						Duration: 30 * time.Minute,
						Comment:  "email / backlog grooming / slack",
					},
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 17, 0, 0,
							0, now.Location()),
						Duration: 15 * time.Minute,
						Comment:  "email / backlog grooming / slack",
					},
				},
				"INTERNAL-2": {
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0,
							0, now.Location()),
						Duration: 70 * time.Minute,
						Comment:  "standup",
					},
				},
				"INTERNAL-3": {
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 13, 10, 0,
							0, now.Location()),
						Duration: 50 * time.Minute,
						Comment:  "primary on-call",
					},
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 14, 35, 0,
							0, now.Location()),
						Duration: 10 * time.Minute,
						Comment:  "primary on-call",
					},
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0,
							0, now.Location()),
						Duration: 15 * time.Minute,
						Comment:  "primary on-call",
					},
				},
				"BAR-1": {
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 14, 45, 0,
							0, now.Location()),
						Duration: 15 * time.Minute,
						Comment:  "bar customer weekly meeting",
					},
				},
				"BAR-2": {
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 15, 50, 0,
							0, now.Location()),
						Duration: 10 * time.Minute,
						Comment:  "check internal tracker ticket re: tls tunnelling",
					},
				},
				"INTERNAL-4": {
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 15, 15, 0,
							0, now.Location()),
						Duration: 15 * time.Minute,
						Comment:  "platform sync",
					},
				},
			},
		},
		"worklog1": {
			input: &parseInput{
				dataFile: "testdata/worklog1",
				config: &config.Config{
					Issues: []config.Issue{
						{
							ID: "ADMIN-1",
							Regexes: wrapRegexes([]string{
								"^admin$",
							}),
						},
					},
					Ignore: wrapRegexes([]string{
						"^lunch$",
					}),
				},
			},
			expect: map[string][]parse.Worklog{
				"ADMIN-1": {
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0,
							0, now.Location()),
						Duration: 45 * time.Minute,
						Comment:  "",
					},
				},
				"XYZ-123": {
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 9, 45, 0,
							0, now.Location()),
						Duration: 135 * time.Minute,
						Comment:  "fighting fires",
					},
				},
				"ABC-987": {
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 13, 0, 0,
							0, now.Location()),
						Duration: 60 * time.Minute,
						Comment:  "more meetings after...\nlunch",
					},
				},
				"ABC-988": {
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 14, 0, 0,
							0, now.Location()),
						Duration: 30 * time.Minute,
						Comment:  "will the meetings\never stop?",
					},
				},
			},
		},
		"worklog2": {
			input: &parseInput{
				dataFile: "testdata/worklog2",
				config: &config.Config{
					Issues: []config.Issue{
						{
							ID: "ADMIN-1",
							Regexes: wrapRegexes([]string{
								"^admin$",
							}),
						},
					},
					Ignore: wrapRegexes([]string{
						"^lunch$",
					}),
				},
			},
			expect: map[string][]parse.Worklog{
				"ADMIN-1": {
					{
						Started: time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0,
							0, now.Location()),
						Duration: 45 * time.Minute,
						Comment:  "",
					},
				},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			f, err := os.Open(tc.input.dataFile)
			if err != nil {
				tt.Fatal(err)
			}
			worklogs, err := parse.Input(f, tc.input.config)
			if err != nil {
				tt.Fatal(err)
			}
			if !reflect.DeepEqual(worklogs, tc.expect) {
				tt.Fatalf("expected:\n%v\n\n---\n\ngot:\n%v", tc.expect, worklogs)
			}
		})
	}
}
