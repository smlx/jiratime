package main

import "fmt"

// VersionCmd represents the `version` command.
type VersionCmd struct{}

// Run the Version command.
func (*VersionCmd) Run() error {
	fmt.Printf("jiratime version %s compiled on %s\n", version, date)
	return nil
}
