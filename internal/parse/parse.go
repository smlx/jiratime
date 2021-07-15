package parse

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/smlx/jiratime/internal/config"
)

var timeRange = regexp.MustCompile(`^[0-9]{4}-[0-9]{4}\s?$`)

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

func identifyIssue(line string, c *config.Config) (string, error) {
	for _, issue := range c.Issues {
		for _, r := range issue.Regexes {
			if r.MatchString(line) {
				return issue.ID, nil
			}
		}
	}
	return "", fmt.Errorf("couldn't match issue to line: %v", line)
}

// Input parses text form stdin and returns an issue-Worklog map.
func Input(r io.Reader, c *config.Config) (map[string][]Worklog, error) {
	var duration time.Duration
	var issue, comment string
	worklogs := map[string][]Worklog{}
	buf := bufio.NewReader(r)
	for line, err := buf.ReadString('\n'); err != io.EOF; line, err = buf.ReadString('\n') {
		if err != nil {
			return nil, fmt.Errorf("couldn't read line: %v", err)
		}
		// strip trailing newline
		line = strings.TrimSpace(line)
		switch {
		case timeRange.MatchString(line):
			if duration == 0 && issue == "" && comment == "" {
				// this is the first time block, continue
			} else if duration == 0 || issue == "" {
				// bad state
				return nil, fmt.Errorf(
					"bad state: duration: %v, issue: %v, comment %v, line: %v",
					duration, issue, comment, line)
			} else {
				// new worklog entry, so add the old Worklog item
				worklogs[issue] = append(worklogs[issue], Worklog{
					Duration: duration,
					Comment:  comment,
				})
			}
			// reset state
			duration = 0
			issue, comment = "", ""
			// parse the time range
			duration, err = parseDuration(line)
			if err != nil {
				return nil, fmt.Errorf("couldn't parse duration: %v", err)
			}
		default:
			if issue == "" {
				// identify an issue, as we don't have one yet
				issue, err = identifyIssue(line, c)
				if err != nil {
					return nil, fmt.Errorf("couldn't identify issue in line `%s`: %v",
						line, err)
				}
			}
			if comment == "" {
				comment = line
			} else {
				comment = strings.Join([]string{comment, line}, "\n")
			}
		}
	}
	// append the final entry
	worklogs[issue] = append(worklogs[issue], Worklog{
		Duration: duration,
		Comment:  comment,
	})
	return worklogs, nil
}
