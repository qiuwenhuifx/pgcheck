package app

import (
	"fmt"
	"strings"

	"github.com/xiongcccc/pgcheck/internal/pgexec"
)

func parseGlobalArgs(args []string) (pgexec.Config, []string, bool, error) {
	cfg := pgexec.DefaultConfig()
	configPath, rest, help, version, err := splitGlobalArgs(args)
	if err != nil {
		return cfg, nil, false, err
	}
	if configPath == "" {
		configPath = pgexec.DefaultConfigPath()
	}
	if err := pgexec.LoadConfigFile(configPath, &cfg); err != nil {
		return cfg, nil, false, err
	}
	cfg.Normalize()
	if err := cfg.Validate(); err != nil {
		return cfg, nil, false, err
	}
	if version {
		return cfg, []string{"version"}, false, nil
	}
	return cfg, rest, help, nil
}

func splitGlobalArgs(args []string) (configPath string, rest []string, help bool, version bool, err error) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--":
			return configPath, append(rest, args[i+1:]...), help, version, nil
		case arg == "--config" || arg == "-c":
			if i+1 >= len(args) {
				return "", nil, false, false, fmt.Errorf("%s requires a config path", arg)
			}
			configPath = args[i+1]
			i++
		case strings.HasPrefix(arg, "--config="):
			configPath = strings.TrimPrefix(arg, "--config=")
		case arg == "--help" || arg == "-h":
			help = true
		case arg == "--version" || arg == "-v":
			version = true
		case strings.HasPrefix(arg, "-"):
			return "", nil, false, false, fmt.Errorf("unknown global option %q; only --config, --help, and --version are supported", arg)
		default:
			rest = append(rest, args[i:]...)
			return configPath, rest, help, version, nil
		}
	}
	return configPath, rest, help, version, nil
}
