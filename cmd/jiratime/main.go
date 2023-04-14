// Package main implements the command-line interface to jiratime.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/alecthomas/kong"
)

// CLI represents the command-line interface.
type CLI struct {
	Submit       SubmitCmd       `kong:"cmd,default=1,help='(default) Submit times'"`
	Authorize    AuthorizeCmd    `kong:"cmd,aliases='auth',help='Get OAuth2 client token'"`
	DumpWorklogs DumpWorklogsCmd `kong:"cmd,help='Dump Worklog records in JSON format'"`
	Version      VersionCmd      `kong:"cmd,help='Print version information'"`
}

// getContext starts a goroutine to handle ^C gracefully, and returns a
// context configured with the given timeout, and a "cancel" function which
// cleans up the signal handling and ensures the goroutine exits. This "cancel"
// function should be deferred in main().
func getContext(timeout time.Duration) (context.Context, func()) {
	ctx, cancel := context.WithDeadline(context.Background(),
		time.Now().Add(timeout))
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		select {
		case <-signalChan:
			log.Println("exiting. ^C again to force.")
			cancel()
		case <-ctx.Done():
		}
		<-signalChan
		os.Exit(130) // https://tldp.org/LDP/abs/html/exitcodes.html
	}()
	return ctx, func() { signal.Stop(signalChan); cancel() }
}

func main() {
	// parse CLI config
	cli := CLI{}
	kctx := kong.Parse(&cli,
		kong.UsageOnError(),
	)
	// execute CLI
	kctx.FatalIfErrorf(kctx.Run())
}
