package parse_test

import (
	"os"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
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
								"^platform ops",
							}),
						},
						{
							ID: "FOO-12",
							Regexes: wrapRegexes([]string{
								"^foo sync",
							}),
						},
						{
							ID: "FOO-3",
							Regexes: wrapRegexes([]string{
								"^fooCustomer devops",
							}),
						},
						{
							ID: "INTERNAL-1",
							Regexes: wrapRegexes([]string{
								"^admin$",
							}),
						},
						{
							ID: "INTERNAL-2",
							Regexes: wrapRegexes([]string{
								"^standup$",
							}),
						},
						{
							ID: "INTERNAL-3",
							Regexes: wrapRegexes([]string{
								"^pd$",
							}),
						},
						{
							ID: "BAR-1",
							Regexes: wrapRegexes([]string{
								"^bar sync",
							}),
						},
						{
							ID: "INTERNAL-4",
							Regexes: wrapRegexes([]string{
								"^platform sync",
							}),
						},
						{
							ID: "BAR-2",
							Regexes: wrapRegexes([]string{
								"^barCustomer infra",
							}),
						},
					},
				},
			},
			expect: map[string][]parse.Worklog{
				"PLATFORM-1": {
					{
						Duration: 20 * time.Minute,
						Comment:  "platform ops",
					},
					{
						Duration: 30 * time.Minute,
						Comment:  "platform ops - example5 cluster melting down again",
					},
				},
				"FOO-12": {
					{
						Duration: 30 * time.Minute,
						Comment:  "foo sync",
					},
				},
				"FOO-3": {
					{
						Duration: 70 * time.Minute,
						Comment:  "fooCustomer devops - node scheduling issue",
					},
					{
						Duration: 15 * time.Minute,
						Comment:  "fooCustomer devops - reply to MS",
					},
				},
				"INTERNAL-1": {
					{
						Duration: 80 * time.Minute,
						Comment:  "admin",
					},
					{
						Duration: 5 * time.Minute,
						Comment:  "admin",
					},
					{
						Duration: 30 * time.Minute,
						Comment:  "admin",
					},
					{
						Duration: 15 * time.Minute,
						Comment:  "admin",
					},
				},
				"INTERNAL-2": {
					{
						Duration: 70 * time.Minute,
						Comment:  "standup",
					},
				},
				"INTERNAL-3": {
					{
						Duration: 50 * time.Minute,
						Comment:  "pd",
					},
					{
						Duration: 10 * time.Minute,
						Comment:  "pd",
					},
					{
						Duration: 15 * time.Minute,
						Comment:  "pd",
					},
				},
				"BAR-1": {
					{
						Duration: 15 * time.Minute,
						Comment:  "bar sync",
					},
				},
				"BAR-2": {
					{
						Duration: 10 * time.Minute,
						Comment:  "barCustomer infra\ncheck internal tracker ticket re: tls tunnelling",
					},
				},
				"INTERNAL-4": {
					{
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
				},
			},
			expect: map[string][]parse.Worklog{
				"ADMIN-1": {
					{
						Duration: 45 * time.Minute,
						Comment:  "admin",
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
			spew.Dump(worklogs)
			spew.Dump(tc.input.config)
			if err != nil {
				tt.Fatal(err)
			}
			if !reflect.DeepEqual(worklogs, tc.expect) {
				tt.Fatalf("expected: %v\n\n---\n\ngot: %v", tc.expect, worklogs)
			}
		})
	}
}
