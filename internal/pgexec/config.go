package pgexec

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Backend    string           `json:"backend"`
	Connection ConnectionConfig `json:"connection"`
	PSQL       PSQLConfig       `json:"psql"`
	Output     OutputConfig     `json:"output"`
}

type ConnectionConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"sslmode"`
}

type PSQLConfig struct {
	Path       string   `json:"path"`
	Quiet      bool     `json:"quiet"`
	TuplesOnly bool     `json:"tuples_only"`
	NoAlign    bool     `json:"no_align"`
	NoPsqlrc   bool     `json:"no_psqlrc"`
	ExtraArgs  []string `json:"extra_args"`
}

type OutputConfig struct {
	Expanded string `json:"expanded"`
}

func DefaultConfig() Config {
	return Config{
		Backend: envDefault("PGCHECK_BACKEND", "psql"),
		Connection: ConnectionConfig{
			Host:     envDefault("PGHOST", "localhost"),
			Port:     envDefault("PGPORT", "5432"),
			User:     os.Getenv("PGUSER"),
			Password: os.Getenv("PGPASSWORD"),
			Database: os.Getenv("PGDATABASE"),
			SSLMode:  envDefault("PGSSLMODE", "disable"),
		},
		PSQL: PSQLConfig{
			Path:     envDefault("PGCHECK_PSQL", "psql"),
			Quiet:    true,
			NoPsqlrc: true,
		},
		Output: OutputConfig{
			Expanded: "auto",
		},
	}
}

func LoadConfigFile(path string, cfg *Config) error {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config %s: %w", path, err)
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("parse config %s: %w", path, err)
	}
	cfg.Normalize()
	return nil
}

func DefaultConfigPath() string {
	if path := os.Getenv("PGCHECK_CONFIG"); path != "" {
		return path
	}
	for _, path := range []string{"pgcheck.json", ".pgcheck.json"} {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	path := filepath.Join(home, ".pgcheck.json")
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

func (c *Config) Normalize() {
	c.Backend = strings.ToLower(strings.TrimSpace(c.Backend))
	if c.Backend == "" {
		c.Backend = "psql"
	}
	c.Output.Expanded = strings.ToLower(strings.TrimSpace(c.Output.Expanded))
	if c.Output.Expanded == "" {
		c.Output.Expanded = "auto"
	}
	if c.Connection.Host == "" {
		c.Connection.Host = "localhost"
	}
	if c.Connection.Port == "" {
		c.Connection.Port = "5432"
	}
	if c.Connection.SSLMode == "" {
		c.Connection.SSLMode = "disable"
	}
	if c.PSQL.Path == "" {
		c.PSQL.Path = "psql"
	}
}

func (c Config) Validate() error {
	switch c.Backend {
	case "psql", "native":
	default:
		return fmt.Errorf("unsupported backend %q", c.Backend)
	}
	switch c.Output.Expanded {
	case "auto", "table", "expanded":
	default:
		return fmt.Errorf("unsupported output.expanded %q; use auto, table, or expanded", c.Output.Expanded)
	}
	if strings.TrimSpace(c.PSQL.Path) == "" {
		return errors.New("psql.path cannot be empty")
	}
	return nil
}

func envDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
