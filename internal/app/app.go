package app

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"github.com/xiongcccc/pgcheck/internal/pgexec"
	"github.com/xiongcccc/pgcheck/internal/queries"
)

type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

type App struct {
	build  BuildInfo
	q      queries.Store
	runner *pgexec.Runner
}

type command struct {
	Name       string
	Usage      string
	Summary    string
	Database   bool
	Extra      []argSpec
	MinVersion int
	Run        func(context.Context, *App, invocation) error
}

type argSpec struct {
	Name     string
	Required bool
}

type invocation struct {
	Command string
	DB      string
	Args    []string
	Version pgexec.ServerVersion
}

func New(sqlFS fs.FS, build BuildInfo) *App {
	cfg := pgexec.DefaultConfig()
	return &App{
		build:  build,
		q:      queries.NewStore(sqlFS),
		runner: pgexec.NewRunner(cfg),
	}
}

func (a *App) Run(args []string) error {
	ctx := context.Background()
	cfg, rest, help, err := parseGlobalArgs(args)
	if err != nil {
		return err
	}
	a.runner = pgexec.NewRunner(cfg)
	if len(rest) == 0 || help || isHelp(rest[0]) {
		a.printHelp()
		return nil
	}
	if rest[0] == "version" {
		fmt.Printf("pgcheck %s (%s, %s)\n", a.build.Version, a.build.Commit, a.build.Date)
		return nil
	}
	if err := a.runner.Check(ctx); err != nil {
		return err
	}

	cmd, ok := commandMap()[strings.ToLower(rest[0])]
	if !ok {
		return fmt.Errorf("unknown command %q; run pgcheck help", rest[0])
	}

	inv := invocation{Command: cmd.Name}
	rest = rest[1:]
	if cmd.Database {
		if len(rest) == 0 {
			return fmt.Errorf("%s requires database name\nusage: %s", cmd.Name, cmd.Usage)
		}
		inv.DB = rest[0]
		rest = rest[1:]
	}
	for _, spec := range cmd.Extra {
		if spec.Required && len(rest) == 0 {
			return fmt.Errorf("%s requires %s\nusage: %s", cmd.Name, spec.Name, cmd.Usage)
		}
	}
	inv.Args = rest

	versionDB := inv.DB
	if versionDB == "" {
		versionDB = cfg.Connection.Database
	}
	version, err := a.runner.ServerVersion(ctx, versionDB)
	if err != nil {
		return fmt.Errorf("detect PostgreSQL server version: %w", err)
	}
	inv.Version = version
	if cmd.MinVersion > 0 && version.Major < cmd.MinVersion {
		return fmt.Errorf("%s requires PostgreSQL %d or later; connected server major version is %d", cmd.Name, cmd.MinVersion, version.Major)
	}
	return cmd.Run(ctx, a, inv)
}

func (a *App) printHelp() {
	fmt.Printf("pgcheck %s - PostgreSQL health check toolkit\n\n", a.build.Version)
	fmt.Println("Usage:")
	fmt.Println("  pgcheck [global options] <command> [database] [arguments]")
	fmt.Println()
	fmt.Println("Connection:")
	fmt.Println("  Use --config, command-line options, or libpq-compatible environment variables.")
	fmt.Println("  CLI options override config files; config files override environment defaults.")
	fmt.Println("  Auto-discovered config files: ./pgcheck.json, ./.pgcheck.json, ~/.pgcheck.json.")
	fmt.Println()
	fmt.Println("Global options:")
	fmt.Println("  --config <path>                              Read JSON config file")
	fmt.Println("  --backend psql|native                        Select execution backend; psql preserves psql formatting, native uses Go driver")
	fmt.Println("  --psql <path>                                psql executable path, same purpose as psql.path in config")
	fmt.Println("  --host/--port/--user/--password/--dbname     Connection settings")
	fmt.Println("  --sslmode <mode>                             PostgreSQL sslmode")
	fmt.Println("  --display auto|table|expanded                Output mode; auto keeps each command's default, expanded is like psql -x")
	fmt.Println("  --expanded / --table                         Shortcuts for display mode")
	fmt.Println("  --quiet / --no-quiet                         Toggle psql -q; quiet=true reduces non-data messages")
	fmt.Println("  --tuples-only                                Same as psql -t; show rows without headers/footers")
	fmt.Println("  --no-align                                   Same as psql -A; use unaligned output")
	fmt.Println("  --no-psqlrc                                  Same as psql -X; ignore ~/.psqlrc for stable output")
	fmt.Println("  --psql-arg <arg>                             Extra raw psql argument; repeatable, e.g. --psql-arg --csv")
	fmt.Println()
	fmt.Println("Config example:")
	fmt.Println("  cp pgcheck.example.json pgcheck.json")
	fmt.Println("  pgcheck --config pgcheck.json dbstatus")
	fmt.Println("  pgcheck --config pgcheck.json --display expanded dbstatus")
	fmt.Println()
	fmt.Println("Connection examples:")
	fmt.Println("  pgcheck --host 127.0.0.1 --port 5432 --user postgres --password secret dbstatus")
	fmt.Println("  pgcheck --backend native --host 127.0.0.1 --user postgres dbstatus")
	fmt.Println("  pgcheck --tuples-only --no-align connections postgres")
	fmt.Println()
	fmt.Println("Commands:")
	cmds := commands()
	sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name < cmds[j].Name })
	for _, cmd := range cmds {
		fmt.Printf("  %-44s %s\n", cmd.Usage, cmd.Summary)
	}
	fmt.Println("  help                                         Show this help")
	fmt.Println("  version                                      Show build version")
}

func commands() []command {
	return []command{
		sqlCommand("alltoast", "pgcheck alltoast <database> <schema>", "List TOAST tables in a schema", true, []argSpec{{"schema", true}}, 0, false, []queryStep{{File: "all_toast.sql", Replacements: []replacement{{Old: "'public'", Arg: 0}}}}),
		{
			Name:    "checkpoint",
			Usage:   "pgcheck checkpoint",
			Summary: "Show background writer and checkpointer statistics",
			Run:     runCheckpoint,
		},
		sqlCommand("connections", "pgcheck connections <database>", "Show connection summary and active queries", true, nil, 0, false, []queryStep{{File: "current_connection.sql"}, {File: "current_query.sql"}}),
		{
			Name: "dbstatus", Usage: "pgcheck dbstatus", Summary: "Show database-level statistics", Run: runDBStatus,
		},
		sqlCommand("freeze", "pgcheck freeze <database>", "Show transaction ID consumption and freeze risk", true, nil, 0, true, []queryStep{{File: "age_consume.sql", Notes: freezeNotes}, {File: "age_consume_dblvl.sql", Header: "Database-level transaction ID usage:"}, {File: "age_consume_rel_lvl.sql", Header: "Top relations by frozen xid age:"}}),
		sqlCommand("index_bloat", "pgcheck index_bloat <database>", "Estimate btree index bloat", true, nil, 0, false, []queryStep{{File: "index_bloat.sql", Header: "This query may take a while.", Notes: indexBloatNotes}}),
		sqlCommand("index_create", "pgcheck index_create <database>", "Show CREATE INDEX progress", true, nil, 12, true, []queryStep{{File: "index_create.sql", Repeat: 5, RepeatDelay: time.Second, EmptyMessage: "no running create index process", Notes: indexCreateNotes}}),
		sqlCommand("index_duplicate", "pgcheck index_duplicate <database>", "Find duplicate indexes", true, nil, 0, false, []queryStep{{File: "index_duplicate.sql"}}),
		sqlCommand("index_low", "pgcheck index_low <database>", "Find low-efficiency indexes", true, nil, 0, false, []queryStep{{File: "index_lower_efficiency.sql", Notes: indexLowNotes}}),
		sqlCommand("index_state", "pgcheck index_state <database>", "Show index details and invalid indexes", true, nil, 0, false, []queryStep{{File: "index_state.sql"}, {File: "index_state_further.sql"}}),
		sqlCommand("lock", "pgcheck lock <database>", "Show lock waits and blocking queue", true, nil, 0, false, []queryStep{{File: "lock_wait_state_further.sql"}, {File: "lock_wait_state.sql", Expanded: true}, {File: "lock_wait_queue.sql", Header: "wait queue:", Notes: lockNotes}}),
		sqlCommand("long_transaction", "pgcheck long_transaction <database>", "Show long-running transactions", true, nil, 0, false, []queryStep{{File: "long_transaction.sql", Notes: longTransactionNotes}}),
		sqlCommand("object", "pgcheck object <database> <user>", "Show objects owned by a user and role membership", true, []argSpec{{"user", true}}, 0, false, []queryStep{{OptionalInline: userObjectSQL, Replacements: []replacement{{Old: "'pgcheck_user'", Arg: 0}}}, {File: "user_member.sql", Header: "user member relationship:"}}),
		sqlCommand("partition", "pgcheck partition <database>", "Show native and inherited partition information", true, nil, 0, false, []queryStep{{File: "partition_info.sql", Header: "native partition:"}, {File: "partition_inherit_info.sql", Header: "inherit and native partition:"}, {File: "partition_size.sql", Header: "partition size:"}}),
		sqlCommand("relation", "pgcheck relation <database> <schema>", "List table and index size in a schema", true, []argSpec{{"schema", true}}, 0, false, []queryStep{{File: "all_relation.sql", Replacements: []replacement{{Old: "'public'", Arg: 0}}}}),
		sqlCommand("relation_bloat", "pgcheck relation_bloat <database>", "Estimate table bloat and vacuum blockers", true, nil, 0, false, []queryStep{{File: "relation_bloat.sql", Header: "This query may take a while. Run ANALYZE first for better estimates.", Notes: relationBloatNotes}, {File: "get_oldest_xmin.sql", Header: "Oldest xmin that may block vacuum:", Expanded: true}, {File: "get_oldest_xact.sql", Header: "Oldest values for vacuum blockers:", Expanded: true}}),
		sqlCommand("relconstraint", "pgcheck relconstraint <database> <relation>", "List constraints and multi-column indexes for a relation", true, []argSpec{{"relation", true}}, 0, false, []queryStep{{File: "rel_constraint.sql", Replacements: []replacement{{Old: "'test'", Arg: 0}}}, {File: "rel_multi_index.sql", Replacements: []replacement{{Old: "'test%'", Arg: 0}}, Notes: relConstraintNotes}}),
		sqlCommand("reltoast", "pgcheck reltoast <database> <relation>", "Show TOAST-related columns for a relation", true, []argSpec{{"relation", true}}, 0, false, []queryStep{{File: "single_toast.sql", Replacements: []replacement{{Old: "'test'", Arg: 0}}}, {File: "single_toast_relation.sql", OptionalInline: singleToastRelationSQL, Replacements: []replacement{{Old: "'test'", Arg: 0}}}}),
		{Name: "replication", Usage: "pgcheck replication", Summary: "Show physical streaming replication status", Run: runReplication},
		sqlCommand("vacuum_need", "pgcheck vacuum_need <database>", "Show tables likely to need vacuum", true, nil, 0, false, []queryStep{{File: "vacuum_need.sql"}}),
		{
			Name:       "vacuum_state",
			Usage:      "pgcheck vacuum_state <database>",
			Summary:    "Show running VACUUM progress",
			Database:   true,
			MinVersion: 11,
			Run:        runVacuumState,
		},
		sqlCommand("wait_event", "pgcheck wait_event <database>", "Show wait events and wait event types", true, nil, 0, false, []queryStep{{File: "wait_event.sql"}}),
		sqlCommand("wal_archive", "pgcheck wal_archive", "Show WAL archiver statistics", false, nil, 0, true, []queryStep{{File: "wal_archive_state.sql"}}),
		sqlCommand("wal_generate", "pgcheck wal_generate <wal_path>", "Show WAL generation speed by scanning pg_wal", false, []argSpec{{"wal_path", true}}, 0, true, []queryStep{{File: "wal_generate_speed.sql", Replacements: []replacement{{Old: "'/home/postgres/15data/pg_wal'", Arg: 0}}}}),
	}
}

func commandMap() map[string]command {
	out := make(map[string]command)
	for _, cmd := range commands() {
		out[cmd.Name] = cmd
	}
	return out
}

func isHelp(s string) bool {
	switch strings.ToLower(s) {
	case "help", "h", "-h", "--help":
		return true
	default:
		return false
	}
}
