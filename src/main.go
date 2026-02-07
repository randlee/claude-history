package main

import (
	"github.com/randlee/claude-history/cmd"
)

// Version information (set by GoReleaser)
var (
	version = "0.3.0-dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersion(version, commit, date)
	cmd.Execute()
}
