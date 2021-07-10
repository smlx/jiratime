package client

import (
	"fmt"

	jira "github.com/andygrunwald/go-jira"
	"github.com/smlx/jiratime/internal/parse"
)

// UploadWorklogs uploads the given worklogs to JIRA, with the
// given day offset (e.g. -1 == yesterday).
func UploadWorklogs(worklogs map[string][]parse.Worklog, dayOffset int) error {

	// TODO check for auth

	client, _ := jira.NewClient(nil, "https://issues.apache.org/jira/")
	issue, _, _ := client.Issue.Get("MESOS-3325", nil)
	fmt.Println(issue)
	return nil
}
