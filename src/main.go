package main

import (
	"github.com/randlee/claude-history/cmd"
	"github.com/randlee/claude-history/pkg/version"
)

// Version information (set by GoReleaser via ldflags)
var (
	versionVar = version.Version // Use constant as default, can be overridden by ldflags
	commit     = "none"
	date       = "unknown"
)

func main() {
	cmd.SetVersion(versionVar, commit, date)
	cmd.Execute()
}
