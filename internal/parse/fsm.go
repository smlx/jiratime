package parse

import (
	"sync"
	"time"

	"github.com/smlx/fsm"
)

const (
	_ fsm.State = iota
	start
	gotDuration
	gotExplicitIssue
	gotImplicitIssue
	end
)

const (
	_ fsm.Event = iota
	matchDuration
	matchExplicitIssue
	noMatch
	eof
)

// A TimesheetParser parses a list of time periods and comments.
// It stores additional state related to timesheet parsing.
type TimesheetParser struct {
	fsm.Machine
	mu sync.Mutex
	// line is the latest line read
	line string
	// started is the parsed start time for the current state
	started time.Time
	// duration is the parsed duration for the current state
	duration time.Duration
	// comment is appended to until worklog submission
	comment []string
	// issue is the JIRA issue name e.g. XYZ-123
	issue string
}

// Occur handles an event occurence.
func (t *TimesheetParser) Occur(e fsm.Event, l string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.line = l
	return t.Machine.Occur(e)
}

// timesheetTransitions defines the transitions for the TimesheetParser FSM
var timesheetTransitions = []fsm.Transition{
	{
		// first entry
		Src:   start,
		Event: matchDuration,
		Dst:   gotDuration,
	}, {
		// first line is XYZ-123...
		Src:   gotDuration,
		Event: matchExplicitIssue,
		Dst:   gotExplicitIssue,
	}, {
		// subsequent lines after first line XYZ-123...
		Src:   gotExplicitIssue,
		Event: noMatch,
		Dst:   gotExplicitIssue,
	}, {
		// attempt match against config
		Src:   gotDuration,
		Event: noMatch,
		Dst:   gotImplicitIssue,
	}, {
		// subsequent lines after config match
		Src:   gotImplicitIssue,
		Event: noMatch,
		Dst:   gotImplicitIssue,
	}, {
		// new entry after config match
		Src:   gotImplicitIssue,
		Event: matchDuration,
		Dst:   gotDuration,
	}, {
		// new entry after explicit issue
		Src:   gotExplicitIssue,
		Event: matchDuration,
		Dst:   gotDuration,
	}, {
		// last entry
		Src:   gotExplicitIssue,
		Event: eof,
		Dst:   end,
	}, {
		// last entry
		Src:   gotImplicitIssue,
		Event: eof,
		Dst:   end,
	},
}
