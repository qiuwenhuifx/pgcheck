package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseGlobalArgsLoadsConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pgcheck.json")
	data := []byte(`{
  "connection": {
    "host": "from-config",
    "port": "5433",
    "user": "postgres",
    "database": "postgres"
  },
  "psql": {
    "path": "psql",
    "quiet": false,
    "no_psqlrc": true
  },
  "output": {
    "expanded": "table"
  }
}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, rest, help, err := parseGlobalArgs([]string{"--config", path, "dbstatus"})
	if err != nil {
		t.Fatal(err)
	}
	if help {
		t.Fatalf("did not expect help")
	}
	if len(rest) != 1 || rest[0] != "dbstatus" {
		t.Fatalf("unexpected remaining args: %v", rest)
	}
	if cfg.Connection.Host != "from-config" {
		t.Fatalf("host = %q", cfg.Connection.Host)
	}
	if cfg.Connection.Port != "5433" {
		t.Fatalf("port = %q", cfg.Connection.Port)
	}
	if cfg.Output.Expanded != "table" {
		t.Fatalf("expanded = %q", cfg.Output.Expanded)
	}
	if cfg.PSQL.Quiet {
		t.Fatalf("quiet should be false")
	}
}

func TestParseGlobalArgsRejectsUnknownOptions(t *testing.T) {
	_, _, _, err := parseGlobalArgs([]string{"--host", "127.0.0.1", "dbstatus"})
	if err == nil {
		t.Fatalf("expected unknown option error")
	}
}
