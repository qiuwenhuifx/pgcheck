package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/xiongcccc/pgcheck/internal/pgexec"
	"github.com/xiongcccc/pgcheck/internal/queries"
)

type replacement struct {
	Old string
	Arg int
}

type queryStep struct {
	File           string
	OptionalInline string
	Header         string
	Notes          []string
	Replacements   []replacement
	Expanded       bool
	Repeat         int
	RepeatDelay    time.Duration
	EmptyMessage   string
}

func sqlCommand(name, usage, summary string, database bool, extra []argSpec, minVersion int, expanded bool, steps []queryStep) command {
	return command{
		Name:       name,
		Usage:      usage,
		Summary:    summary,
		Database:   database,
		Extra:      extra,
		MinVersion: minVersion,
		Run: func(ctx context.Context, a *App, inv invocation) error {
			for _, step := range steps {
				if !step.Expanded {
					step.Expanded = expanded
				}
				if err := a.runStep(ctx, inv, step); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func (a *App) runStep(ctx context.Context, inv invocation, step queryStep) error {
	query, err := a.stepSQL(step)
	if err != nil {
		return err
	}
	for _, repl := range step.Replacements {
		if repl.Arg >= len(inv.Args) {
			return fmt.Errorf("missing argument for %s", step.File)
		}
		query = queries.ReplaceLiteral(query, repl.Old, inv.Args[repl.Arg])
	}
	if step.Header != "" {
		fmt.Println(step.Header)
	}
	opts := pgexec.Options{Database: inv.DB, Expanded: a.displayExpanded(step.Expanded)}
	repeat := step.Repeat
	if repeat == 0 {
		repeat = 1
	}
	if step.EmptyMessage != "" {
		hasRows, err := a.hasResultRows(ctx, opts, query)
		if err != nil {
			return err
		}
		if !hasRows {
			fmt.Println(step.EmptyMessage)
			return nil
		}
	}
	for i := 0; i < repeat; i++ {
		if err := a.runner.Exec(ctx, opts, query); err != nil {
			return err
		}
		if i < repeat-1 && step.RepeatDelay > 0 {
			time.Sleep(step.RepeatDelay)
		}
	}
	printNotes(step.Notes)
	return nil
}

func (a *App) stepSQL(step queryStep) (string, error) {
	if step.OptionalInline != "" {
		return strings.TrimSpace(step.OptionalInline), nil
	}
	return a.q.Read(step.File)
}

func (a *App) hasResultRows(ctx context.Context, opts pgexec.Options, query string) (bool, error) {
	wrapped := "SELECT EXISTS (" + strings.TrimRight(query, " \t\r\n;") + ")"
	out, err := a.runner.QueryScalar(ctx, pgexec.Options{Database: opts.Database, TuplesOnly: true, NoAlign: true}, wrapped)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) == "t", nil
}

func printNotes(notes []string) {
	if len(notes) == 0 {
		return
	}
	fmt.Println("- Description:")
	for i, note := range notes {
		fmt.Printf("%d. %s\n", i+1, note)
	}
}

func runReplication(ctx context.Context, a *App, inv invocation) error {
	file := "replication_primary_state_10.sql"
	if inv.Version.Major >= 13 {
		file = "replication_primary_state_13.sql"
	}
	query, err := a.q.Read(file)
	if err != nil {
		return err
	}
	return a.runner.Exec(ctx, pgexec.Options{Expanded: a.displayExpanded(false)}, query)
}

func runDBStatus(ctx context.Context, a *App, inv invocation) error {
	query := databaseStatusSQL(inv.Version.Major)
	return a.runner.Exec(ctx, pgexec.Options{Expanded: a.displayExpanded(true)}, query)
}

func runCheckpoint(ctx context.Context, a *App, inv invocation) error {
	file := "bgworker_checkpoint_state.sql"
	if inv.Version.Major >= 17 {
		file = "bgworker_checkpoint_state_17.sql"
	}
	query, err := a.q.Read(file)
	if err != nil {
		return err
	}
	if err := a.runner.Exec(ctx, pgexec.Options{Expanded: a.displayExpanded(true)}, query); err != nil {
		return err
	}
	printNotes(checkpointNotes)
	return nil
}

func runVacuumState(ctx context.Context, a *App, inv invocation) error {
	file := "vacuum_state.sql"
	if inv.Version.Major >= 17 {
		file = "vacuum_state_17.sql"
	}
	step := queryStep{
		File:         file,
		Expanded:     true,
		Repeat:       5,
		RepeatDelay:  time.Second,
		EmptyMessage: "no running vacuum process",
		Notes:        vacuumStateNotes,
	}
	return a.runStep(ctx, inv, step)
}

func (a *App) displayExpanded(defaultExpanded bool) bool {
	switch a.runner.Config.Output.Expanded {
	case "expanded":
		return true
	case "table":
		return false
	default:
		return defaultExpanded
	}
}

func databaseStatusSQL(major int) string {
	base := []string{
		"datname AS database_name",
		"pg_size_pretty(pg_database_size(datname)) AS database_size",
		"100 * blks_hit / NULLIF(blks_hit + blks_read, 0) || ' %' AS cache_hit_ratio",
		"100 * xact_commit / NULLIF(xact_commit + xact_rollback, 0) || ' %' AS commit_ratio",
		"conflicts",
		"temp_files",
		"pg_size_pretty(temp_bytes) AS temp_bytes",
		"deadlocks",
	}
	if major >= 12 {
		base = append(base, "checksum_failures")
	}
	base = append(base, "blk_read_time", "blk_write_time")
	if major >= 14 {
		base = append(base,
			"session_time",
			"active_time",
			"idle_in_transaction_time",
			"sessions",
			"sessions_abandoned",
			"sessions_fatal",
			"sessions_killed",
		)
	}
	return "SELECT\n    " + strings.Join(base, ",\n    ") + "\nFROM pg_stat_database\nWHERE (blks_hit + blks_read) > 0\n  AND datname NOT LIKE '%template%';"
}

const singleToastRelationSQL = `
SELECT
    n.nspname as schema,
    s.oid::regclass as relname,
    s.reltoastrelid::regclass as toast_name,
    pg_relation_size(s.reltoastrelid) AS toast_size
FROM pg_class s
JOIN pg_namespace n ON s.relnamespace = n.oid
WHERE relname = 'test'
  AND reltoastrelid <> 0
ORDER BY 3 DESC;`

const userObjectSQL = `
SELECT
    nsp.nspname AS schema_name,
    cls.relname AS object_name,
    rol.rolname AS object_owner,
    CASE cls.relkind
        WHEN 'r' THEN 'TABLE'
        WHEN 'p' THEN 'PARTITIONED_TABLE'
        WHEN 'm' THEN 'MATERIALIZED_VIEW'
        WHEN 'i' THEN 'INDEX'
        WHEN 'I' THEN 'PARTITIONED_INDEX'
        WHEN 'S' THEN 'SEQUENCE'
        WHEN 'v' THEN 'VIEW'
        WHEN 'c' THEN 'TYPE'
        ELSE cls.relkind::text
    END AS object_type
FROM pg_class cls
JOIN pg_roles rol ON rol.oid = cls.relowner
JOIN pg_namespace nsp ON nsp.oid = cls.relnamespace
WHERE nsp.nspname NOT IN ('information_schema', 'pg_catalog')
  AND nsp.nspname NOT LIKE 'pg_toast%'
  AND rol.rolname = 'pgcheck_user'
ORDER BY nsp.nspname, cls.relname;`
