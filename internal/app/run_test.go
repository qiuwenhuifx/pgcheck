package app

import (
	"strings"
	"testing"
)

func TestDatabaseStatusSQLVersionColumns(t *testing.T) {
	pg11 := databaseStatusSQL(11)
	if strings.Contains(pg11, "checksum_failures") {
		t.Fatalf("PostgreSQL 11 query should not include checksum_failures")
	}
	if strings.Contains(pg11, "session_time") {
		t.Fatalf("PostgreSQL 11 query should not include session_time")
	}

	pg12 := databaseStatusSQL(12)
	if !strings.Contains(pg12, "checksum_failures") {
		t.Fatalf("PostgreSQL 12 query should include checksum_failures")
	}
	if strings.Contains(pg12, "session_time") {
		t.Fatalf("PostgreSQL 12 query should not include session_time")
	}

	pg14 := databaseStatusSQL(14)
	if !strings.Contains(pg14, "checksum_failures") || !strings.Contains(pg14, "session_time") {
		t.Fatalf("PostgreSQL 14 query should include checksum and session statistics")
	}
}

func TestCommandRegistryHasUniqueNames(t *testing.T) {
	seen := map[string]bool{}
	for _, cmd := range commands() {
		if cmd.Name == "" {
			t.Fatalf("command name cannot be empty")
		}
		if cmd.Usage == "" {
			t.Fatalf("%s usage cannot be empty", cmd.Name)
		}
		if seen[cmd.Name] {
			t.Fatalf("duplicate command name %q", cmd.Name)
		}
		seen[cmd.Name] = true
	}
}
