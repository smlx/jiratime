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
	ignore
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
	// defaultComment is appended to comment if comment is otherwise empty
	defaultComment string
	// issue is the Jira issue name e.g. XYZ-123
	issue string
}

// Occur handles an event occurrence.
func (t *TimesheetParser) Occur(e fsm.Event, l string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.line = l
	return t.Machine.Occur(e)
}

// timesheetTransitions defines the transitions for the TimesheetParser FSM
var timesheetTransitions = []fsm.Transition{
	{
		// first entry, or after ignore
		Src:   start,
		Event: matchDuration,
		Dst:   gotDuration,
	}, {
		// first line is XYZ-123, explicitly identifying an issue
		Src:   gotDuration,
		Event: matchExplicitIssue,
		Dst:   gotExplicitIssue,
	}, {
		// subsequent lines after explicit issue are comments until the next
		// duration is found
		Src:   gotExplicitIssue,
		Event: noMatch,
		Dst:   gotExplicitIssue,
	}, {
		Src:   gotExplicitIssue,
		Event: ignore,
		Dst:   gotExplicitIssue,
	}, {
		// match first line of timesheet entry against config
		Src:   gotDuration,
		Event: noMatch,
		Dst:   gotImplicitIssue,
	}, {
		// subsequent lines after config match are comments until the next duration
		// is found
		Src:   gotImplicitIssue,
		Event: noMatch,
		Dst:   gotImplicitIssue,
	}, {
		Src:   gotImplicitIssue,
		Event: ignore,
		Dst:   gotImplicitIssue,
	}, {
		// issue hasn't been identified and match an ignore regex: return to start
		Src:   gotDuration,
		Event: ignore,
		Dst:   start,
	}, {
		// new duration signifies a new timesheet entry
		Src:   gotImplicitIssue,
		Event: matchDuration,
		Dst:   gotDuration,
	}, {
		Src:   gotExplicitIssue,
		Event: matchDuration,
		Dst:   gotDuration,
	}, {
		// reached the end of the timesheet
		Src:   gotExplicitIssue,
		Event: eof,
		Dst:   end,
	}, {
		Src:   gotImplicitIssue,
		Event: eof,
		Dst:   end,
	}, {
		Src:   start,
		Event: eof,
		Dst:   end,
	},
}
