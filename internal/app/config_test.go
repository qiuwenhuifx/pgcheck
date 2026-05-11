package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseGlobalArgsOverridesConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pgcheck.json")
	data := []byte(`{
  "backend": "psql",
  "connection": {
    "host": "from-config",
    "port": "5432",
    "user": "postgres",
    "database": "postgres"
  },
  "psql": {
    "path": "psql",
    "quiet": true,
    "no_psqlrc": true
  },
  "output": {
    "expanded": "auto"
  }
}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, rest, help, err := parseGlobalArgs([]string{
		"--config", path,
		"--backend", "native",
		"--host", "from-flag",
		"--display", "expanded",
		"--no-quiet",
		"dbstatus",
	})
	if err != nil {
		t.Fatal(err)
	}
	if help {
		t.Fatalf("did not expect help")
	}
	if len(rest) != 1 || rest[0] != "dbstatus" {
		t.Fatalf("unexpected remaining args: %v", rest)
	}
	if cfg.Backend != "native" {
		t.Fatalf("backend = %q", cfg.Backend)
	}
	if cfg.Connection.Host != "from-flag" {
		t.Fatalf("host = %q", cfg.Connection.Host)
	}
	if cfg.Output.Expanded != "expanded" {
		t.Fatalf("expanded = %q", cfg.Output.Expanded)
	}
	if cfg.PSQL.Quiet {
		t.Fatalf("quiet should be false")
	}
}
