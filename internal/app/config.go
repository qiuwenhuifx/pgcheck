package app

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/xiongcccc/pgcheck/internal/pgexec"
)

type stringList []string

func (s *stringList) String() string {
	return strings.Join(*s, ",")
}

func (s *stringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func parseGlobalArgs(args []string) (pgexec.Config, []string, bool, error) {
	cfg := pgexec.DefaultConfig()
	configPath := findConfigPath(args)
	if configPath == "" {
		configPath = pgexec.DefaultConfigPath()
	}
	if err := pgexec.LoadConfigFile(configPath, &cfg); err != nil {
		return cfg, nil, false, err
	}

	extraArgs := stringList(append([]string{}, cfg.PSQL.ExtraArgs...))
	fs := flag.NewFlagSet("pgcheck", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	help := fs.Bool("help", false, "show help")
	fs.BoolVar(help, "h", false, "show help")
	version := fs.Bool("version", false, "show version")
	fs.BoolVar(version, "v", false, "show version")
	fs.StringVar(&configPath, "config", configPath, "config file path")
	fs.StringVar(&cfg.Backend, "backend", cfg.Backend, "execution backend: psql or native")
	fs.StringVar(&cfg.PSQL.Path, "psql", cfg.PSQL.Path, "psql executable path")
	fs.BoolVar(&cfg.PSQL.Quiet, "quiet", cfg.PSQL.Quiet, "pass -q to psql")
	noQuiet := fs.Bool("no-quiet", false, "do not pass -q to psql")
	fs.BoolVar(&cfg.PSQL.NoPsqlrc, "no-psqlrc", cfg.PSQL.NoPsqlrc, "pass -X to psql")
	fs.BoolVar(&cfg.PSQL.TuplesOnly, "tuples-only", cfg.PSQL.TuplesOnly, "pass -t to psql and suppress headers in native output")
	fs.BoolVar(&cfg.PSQL.NoAlign, "no-align", cfg.PSQL.NoAlign, "pass -A to psql and use unaligned native output")
	fs.Var(&extraArgs, "psql-arg", "extra raw psql argument; may be repeated")
	fs.StringVar(&cfg.Output.Expanded, "display", cfg.Output.Expanded, "display mode: auto, table, or expanded")
	expanded := fs.Bool("expanded", false, "force expanded output")
	table := fs.Bool("table", false, "force table output")
	fs.StringVar(&cfg.Connection.Host, "host", cfg.Connection.Host, "PostgreSQL host")
	fs.StringVar(&cfg.Connection.Port, "port", cfg.Connection.Port, "PostgreSQL port")
	fs.StringVar(&cfg.Connection.User, "user", cfg.Connection.User, "PostgreSQL user")
	fs.StringVar(&cfg.Connection.Password, "password", cfg.Connection.Password, "PostgreSQL password")
	fs.StringVar(&cfg.Connection.Database, "dbname", cfg.Connection.Database, "default PostgreSQL database")
	fs.StringVar(&cfg.Connection.SSLMode, "sslmode", cfg.Connection.SSLMode, "PostgreSQL sslmode")
	if err := fs.Parse(args); err != nil {
		return cfg, nil, false, fmt.Errorf("%w; run pgcheck help", err)
	}
	cfg.PSQL.ExtraArgs = []string(extraArgs)
	if *noQuiet {
		cfg.PSQL.Quiet = false
	}
	if *expanded {
		cfg.Output.Expanded = "expanded"
	}
	if *table {
		cfg.Output.Expanded = "table"
	}
	cfg.Normalize()
	if err := cfg.Validate(); err != nil {
		return cfg, nil, false, err
	}
	if *version {
		return cfg, []string{"version"}, false, nil
	}
	return cfg, fs.Args(), *help, nil
}

func findConfigPath(args []string) string {
	for i, arg := range args {
		if arg == "--" {
			return ""
		}
		if arg == "--config" && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(arg, "--config=") {
			return strings.TrimPrefix(arg, "--config=")
		}
	}
	return ""
}
