package parse

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/smlx/fsm"
	"github.com/smlx/jiratime/internal/config"
)

var timeRange = regexp.MustCompile(`^[0-9]{4}-[0-9]{4}\s?$`)
var jiraIssue = regexp.MustCompile(`^([A-Za-z]+-[0-9]+)(\s.+)?$`)

// Worklog represents an individual work log entry on a ticket.
type Worklog struct {
	Duration time.Duration
	Comment  string // optional
}

func parseDuration(t string) (time.Duration, error) {
	times := strings.Split(strings.TrimSpace(t), "-")
	if len(times) != 2 {
		return 0, fmt.Errorf("bad timeRange format")
	}
	start, err := time.Parse("1504", times[0])
	if err != nil {
		return 0, fmt.Errorf("couldn't parse start time: %v", err)
	}
	end, err := time.Parse("1504", times[1])
	if err != nil {
		return 0, fmt.Errorf("couldn't parse end time: %v", err)
	}
	duration := end.Sub(start)
	if duration <= 0 {
		return 0, fmt.Errorf("invalid duration, less than 1 minute")
	}
	return duration, nil
}

func getImplicitIssue(line string, c *config.Config) (string, string, string, error) {
	for _, issue := range c.Issues {
		for _, r := range issue.Regexes {
			if matches := r.FindStringSubmatch(line); matches != nil {
				var comment string
				if len(matches) > 1 {
					comment = strings.Trim(matches[1], " -")
				}
				return issue.ID, issue.DefaultComment, comment, nil
			}
		}
	}
	return "", "", "", fmt.Errorf("couldn't match issue to line: %v", line)
}

func appendWorklog() error {
	// TODO: deduplicate lines here
	return nil
}

// Input parses text form stdin and returns an issue-Worklog map.
func Input(r io.Reader, c *config.Config) (map[string][]Worklog, error) {
	var err error
	worklogs := map[string][]Worklog{}
	buf := bufio.NewReader(r)

	timesheet := TimesheetParser{
		Machine: fsm.Machine{
			State:       start, // initial state
			Transitions: timesheetTransitions,
		},
	}

	timesheet.OnEntry = map[fsm.State][]fsm.TransitionFunc{
		gotDuration: {
			func(_ fsm.Event) error {
				// If we are transitioning from start then this is the first entry and
				// there is nothing to submit yet.
				if timesheet.State != start {
					worklogs[timesheet.issue] = append(worklogs[timesheet.issue], Worklog{
						Duration: timesheet.duration,
						Comment:  strings.Join(timesheet.comment, "\n"),
					})
				}
				// set the duration
				timesheet.duration, err = parseDuration(timesheet.line)
				return err
			},
		},
		gotExplicitIssue: {
			func(e fsm.Event) error {
				if e == noMatch {
					timesheet.comment =
						append(timesheet.comment, strings.Trim(timesheet.line, " -"))
					return nil
				}
				// we have just identified an explicit issue on the first line of an
				// entry, so reset state
				matches := jiraIssue.FindStringSubmatch(timesheet.line)
				timesheet.issue = matches[1]
				if matches[2] == "" {
					timesheet.comment = nil
				} else {
					timesheet.comment = []string{strings.Trim(matches[2], " -")}
				}
				return nil
			},
		},
		gotImplicitIssue: {
			func(e fsm.Event) error {
				if timesheet.State == gotDuration {
					// we haven't identified an issue yet, so try to do so here
					var defaultComment, comment string
					timesheet.issue, defaultComment, comment, err =
						getImplicitIssue(timesheet.line, c)
					timesheet.comment = nil
					if defaultComment != "" {
						timesheet.comment = append(timesheet.comment, defaultComment)
					}
					if comment != "" {
						timesheet.comment = append(timesheet.comment, comment)
					}
					return err
				}
				// we are just appending comments here
				timesheet.comment = append(timesheet.comment, timesheet.line)
				return nil
			},
		},
		end: {
			func(e fsm.Event) error {
				// insert the final entry
				worklogs[timesheet.issue] = append(worklogs[timesheet.issue], Worklog{
					Duration: timesheet.duration,
					Comment:  strings.Join(timesheet.comment, "\n"),
				})
				return nil
			},
		},
	}

	for line, err := buf.ReadString('\n'); err != io.EOF; line, err = buf.ReadString('\n') {
		if err != nil {
			return nil, fmt.Errorf("couldn't read line: %v", err)
		}
		line = strings.TrimSpace(line) // strip trailing newline
		switch {
		case timeRange.MatchString(line):
			if err = timesheet.Occur(matchDuration, line); err != nil {
				return nil, err
			}
		case jiraIssue.MatchString(line):
			if err = timesheet.Occur(matchExplicitIssue, line); err != nil {
				return nil, err
			}
		default:
			if err = timesheet.Occur(noMatch, line); err != nil {
				return nil, err
			}
		}
	}
	if err = timesheet.Machine.Occur(eof); err != nil {
		return nil, err
	}
	return worklogs, nil
}
