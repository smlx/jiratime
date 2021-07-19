# JIRAtime

![Tag and Release](https://github.com/smlx/jiratime/workflows/Tag%20and%20Release/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/smlx/jiratime/badge.svg?branch=main)](https://coveralls.io/github/smlx/jiratime?branch=main)

`jiratime` makes it easy to submit timesheets to JIRA quickly from the command line.
It is designed for use with timesheets logged in (neo)vim.

## Get it

Download the latest [release](https://github.com/smlx/jiratime/releases) on github, or:

```
go install github.com/smlx/jiratime/cmd/jiratime@latest
```

## Configure it

`jiratime` reads configuration from `$XDG_CONFIG_HOME/jiratime/config.yml`

## Use it

### Timesheet format

The timesheet format is minimal and opinionated.

```
TODO: documentation
```

### Authorization

`jiratime` requires a one-time authorization in JIRA cloud.

```
TODO: documentation
```

### Timesheet submission

With no command specified or with `submit` specified, `jiratime` performs timesheet submission by reading from STDIN.

```
TODO: documentation and a GIF
```

### Options

Run `jiratime help` to discover the command line options.
