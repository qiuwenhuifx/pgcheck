package app

import (
	"fmt"
	"strings"

	"github.com/xiongcccc/pgcheck/internal/pgexec"
)

func parseGlobalArgs(args []string) (pgexec.Config, []string, bool, error) {
	cfg := pgexec.DefaultConfig()
	configPath, overrides, rest, help, version, err := splitGlobalArgs(args)
	if err != nil {
		return cfg, nil, false, err
	}
	if configPath == "" {
		configPath = pgexec.DefaultConfigPath()
	}
	if err := pgexec.LoadConfigFile(configPath, &cfg); err != nil {
		return cfg, nil, false, err
	}
	overrides.apply(&cfg)
	cfg.Normalize()
	if err := cfg.Validate(); err != nil {
		return cfg, nil, false, err
	}
	if version {
		return cfg, []string{"version"}, false, nil
	}
	return cfg, rest, help, nil
}

type cliOverrides struct {
	expanded   bool
	tuplesOnly bool
	noAlign    bool
}

func (o cliOverrides) apply(cfg *pgexec.Config) {
	if o.expanded {
		cfg.Output.Expanded = "expanded"
	}
	if o.tuplesOnly {
		cfg.PSQL.TuplesOnly = true
	}
	if o.noAlign {
		cfg.PSQL.NoAlign = true
	}
}

func splitGlobalArgs(args []string) (configPath string, overrides cliOverrides, rest []string, help bool, version bool, err error) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--":
			return configPath, overrides, append(rest, args[i+1:]...), help, version, nil
		case arg == "--config" || arg == "-c":
			if i+1 >= len(args) {
				return "", overrides, nil, false, false, fmt.Errorf("%s requires a config path", arg)
			}
			configPath = args[i+1]
			i++
		case strings.HasPrefix(arg, "--config="):
			configPath = strings.TrimPrefix(arg, "--config=")
		case arg == "--help" || arg == "-h":
			help = true
		case arg == "--version" || arg == "-v":
			version = true
		case isOutputShortFlags(arg):
			overrides.applyShortFlags(arg)
		case strings.HasPrefix(arg, "-"):
			return "", overrides, nil, false, false, fmt.Errorf("unknown global option %q; supported global options: --config, --help, --version, -x, -A, -t", arg)
		default:
			rest = append(rest, args[i:]...)
			return configPath, overrides, rest, help, version, nil
		}
	}
	return configPath, overrides, rest, help, version, nil
}

func isOutputShortFlags(arg string) bool {
	if !strings.HasPrefix(arg, "-") || len(arg) < 2 {
		return false
	}
	for _, flag := range strings.TrimPrefix(arg, "-") {
		switch flag {
		case 'x', 'A', 't':
		default:
			return false
		}
	}
	return true
}

func (o *cliOverrides) applyShortFlags(arg string) {
	for _, flag := range strings.TrimPrefix(arg, "-") {
		switch flag {
		case 'x':
			o.expanded = true
		case 'A':
			o.noAlign = true
		case 't':
			o.tuplesOnly = true
		}
	}
}
