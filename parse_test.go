package main_test

import (
	"os"
	"reflect"
	"regexp"
	"testing"
	"time"

	main "github.com/smlx/jiratime"
)

type parseInput struct {
	dataFile string
	config   *main.Config
}

func wrapRegexes(regexes []string) []main.Regexp {
	var regexps []main.Regexp
	for _, s := range regexes {
		r := regexp.MustCompile(s)
		regexps = append(regexps, main.Regexp{*r})
	}
	return regexps
}

func TestParseInput(t *testing.T) {
	var testCases = map[string]struct {
		input  *parseInput
		expect map[string][]main.Worklog
	}{
		"worklog0": {
			input: &parseInput{
				dataFile: "testdata/worklog0",
				config: &main.Config{
					Issues: []main.Issue{
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
			expect: map[string][]main.Worklog{
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
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			f, err := os.Open(tc.input.dataFile)
			if err != nil {
				tt.Fatal(err)
			}
			worklog, err := main.ParseInput(f, tc.input.config)
			if err != nil {
				tt.Fatal(err)
			}
			if !reflect.DeepEqual(worklog, tc.expect) {
				tt.Fatalf("expected: %v\n\n---\n\ngot: %v", tc.expect, worklog)
			}
		})
	}
}
