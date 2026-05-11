package app

import (
	"fmt"
	"runtime"
	"strings"
)

func (b BuildInfo) String() string {
	version := strings.TrimSpace(b.Version)
	if version == "" {
		version = "unknown"
	}
	lines := []string{fmt.Sprintf("pgcheck version %s", version)}
	if isBuildValueSet(b.Commit) {
		lines = append(lines, "commit: "+b.Commit)
	}
	if isBuildValueSet(b.Date) {
		lines = append(lines, "built: "+b.Date)
	}
	lines = append(lines, "go: "+runtime.Version())
	return strings.Join(lines, "\n")
}

func isBuildValueSet(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && value != "dev" && value != "unknown"
}
