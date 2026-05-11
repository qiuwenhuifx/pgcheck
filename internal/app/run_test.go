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

func TestCommandAliases(t *testing.T) {
	aliases := map[string]string{
		"analyze_need":    "analyze_needed",
		"index_low":       "index_efficiency",
		"index_null_frac": "index_null_risk",
		"index_state":     "index_health",
		"int_pk_risk":     "integer_pk_risk",
		"sequence_risk":   "integer_pk_risk",
		"relation_bloat":  "table_bloat",
		"vacuum_need":     "vacuum_needed",
		"xid_wraparound":  "wraparound_risk",
		"xmin_horizon":    "xmin_blockers",
	}
	registry := commandMap()
	for alias, target := range aliases {
		cmd, ok := registry[alias]
		if !ok {
			t.Fatalf("alias %q is missing", alias)
		}
		if cmd.Name != target {
			t.Fatalf("alias %q resolved to %q, want %q", alias, cmd.Name, target)
		}
	}
}

func TestSplitTailFlags(t *testing.T) {
	positional, flags := splitTailFlags([]string{"postgres", "--show-sql", "public", "--explain"})
	if strings.Join(positional, ",") != "postgres,public" {
		t.Fatalf("positional = %v", positional)
	}
	if strings.Join(flags, ",") != "--show-sql,--explain" {
		t.Fatalf("flags = %v", flags)
	}
}
