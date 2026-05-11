package pgexec

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

type Runner struct {
	Bin     string
	Backend string
	Out     io.Writer
	Err     io.Writer
}

type Options struct {
	Database   string
	Expanded   bool
	TuplesOnly bool
}

type ServerVersion struct {
	Num   int
	Major int
	Text  string
}

func NewRunner() *Runner {
	bin := os.Getenv("PGCHECK_PSQL")
	if bin == "" {
		bin = "psql"
	}
	backend := strings.ToLower(os.Getenv("PGCHECK_BACKEND"))
	if backend == "" {
		backend = "psql"
	}
	return &Runner{
		Bin:     bin,
		Backend: backend,
		Out:     os.Stdout,
		Err:     os.Stderr,
	}
}

func (r *Runner) Check(ctx context.Context) error {
	if r.Backend == "native" {
		db, err := sql.Open("postgres", connString(""))
		if err != nil {
			return err
		}
		defer db.Close()
		return db.PingContext(ctx)
	}
	_, err := exec.LookPath(r.Bin)
	if err != nil {
		return fmt.Errorf("cannot find %q in PATH; please install PostgreSQL client tools or set PATH", r.Bin)
	}
	cmd := exec.CommandContext(ctx, r.Bin, "--version")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot execute %q: %w %s", r.Bin, err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func (r *Runner) ServerVersion(ctx context.Context, database string) (ServerVersion, error) {
	out, err := r.QueryScalar(ctx, Options{Database: database, TuplesOnly: true}, "SHOW server_version_num;")
	if err != nil {
		return ServerVersion{}, err
	}
	text := strings.TrimSpace(out)
	num, err := strconv.Atoi(text)
	if err != nil {
		return ServerVersion{}, fmt.Errorf("unexpected server_version_num %q", text)
	}
	return ServerVersion{
		Num:   num,
		Major: num / 10000,
		Text:  text,
	}, nil
}

func (r *Runner) QueryScalar(ctx context.Context, opts Options, query string) (string, error) {
	if r.Backend == "native" {
		db, err := sql.Open("postgres", connString(opts.Database))
		if err != nil {
			return "", err
		}
		defer db.Close()
		var value sql.NullString
		if err := db.QueryRowContext(ctx, query).Scan(&value); err != nil {
			return "", err
		}
		return value.String, nil
	}
	var stdout, stderr bytes.Buffer
	cmd := r.command(ctx, opts, query)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", formatCommandError(err, stderr.String())
	}
	return stdout.String(), nil
}

func (r *Runner) Exec(ctx context.Context, opts Options, query string) error {
	if r.Backend == "native" {
		db, err := sql.Open("postgres", connString(opts.Database))
		if err != nil {
			return err
		}
		defer db.Close()
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			return err
		}
		defer rows.Close()
		return printRows(r.Out, rows, opts.Expanded)
	}
	cmd := r.command(ctx, opts, query)
	cmd.Stdout = r.Out
	cmd.Stderr = r.Err
	if err := cmd.Run(); err != nil {
		return formatCommandError(err, "")
	}
	return nil
}

func connString(database string) string {
	fields := map[string]string{
		"host":     envDefault("PGHOST", "localhost"),
		"port":     envDefault("PGPORT", "5432"),
		"user":     os.Getenv("PGUSER"),
		"password": os.Getenv("PGPASSWORD"),
		"dbname":   database,
		"sslmode":  envDefault("PGSSLMODE", "disable"),
	}
	if fields["dbname"] == "" {
		fields["dbname"] = os.Getenv("PGDATABASE")
	}
	parts := make([]string, 0, len(fields))
	for _, key := range []string{"host", "port", "user", "password", "dbname", "sslmode"} {
		if fields[key] == "" {
			continue
		}
		parts = append(parts, key+"="+quoteConnValue(fields[key]))
	}
	return strings.Join(parts, " ")
}

func envDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func quoteConnValue(value string) string {
	return "'" + strings.ReplaceAll(strings.ReplaceAll(value, `\`, `\\`), `'`, `\'`) + "'"
}

func printRows(out io.Writer, rows *sql.Rows, expanded bool) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]sql.NullString, len(cols))
	dest := make([]any, len(cols))
	for i := range values {
		dest[i] = &values[i]
	}
	widths := make([]int, len(cols))
	for i, col := range cols {
		widths[i] = len(col)
	}
	var data [][]string
	for rows.Next() {
		if err := rows.Scan(dest...); err != nil {
			return err
		}
		row := make([]string, len(cols))
		for i, value := range values {
			if value.Valid {
				row[i] = value.String
			} else {
				row[i] = ""
			}
			if len(row[i]) > widths[i] {
				widths[i] = len(row[i])
			}
		}
		data = append(data, row)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if expanded {
		return printExpanded(out, cols, data)
	}
	return printTable(out, cols, widths, data)
}

func printTable(out io.Writer, cols []string, widths []int, data [][]string) error {
	for i, col := range cols {
		if i > 0 {
			fmt.Fprint(out, "  ")
		}
		fmt.Fprintf(out, "%-*s", widths[i], col)
	}
	fmt.Fprintln(out)
	for i := range cols {
		if i > 0 {
			fmt.Fprint(out, "  ")
		}
		fmt.Fprint(out, strings.Repeat("-", widths[i]))
	}
	fmt.Fprintln(out)
	for _, row := range data {
		for i, cell := range row {
			if i > 0 {
				fmt.Fprint(out, "  ")
			}
			fmt.Fprintf(out, "%-*s", widths[i], cell)
		}
		fmt.Fprintln(out)
	}
	fmt.Fprintf(out, "(%d rows)\n", len(data))
	return nil
}

func printExpanded(out io.Writer, cols []string, data [][]string) error {
	for idx, row := range data {
		fmt.Fprintf(out, "-[ RECORD %d ]-------------------------\n", idx+1)
		for i, col := range cols {
			fmt.Fprintf(out, "%s | %s\n", col, row[i])
		}
	}
	fmt.Fprintf(out, "(%d rows)\n", len(data))
	return nil
}

func (r *Runner) command(ctx context.Context, opts Options, query string) *exec.Cmd {
	args := []string{"-q", "-v", "ON_ERROR_STOP=1"}
	if opts.Expanded {
		args = append(args, "-x")
	}
	if opts.TuplesOnly {
		args = append(args, "-A", "-t")
	}
	if opts.Database != "" {
		args = append(args, "-d", opts.Database)
	}
	args = append(args, "-c", query)
	return exec.CommandContext(ctx, r.Bin, args...)
}

func formatCommandError(err error, stderr string) error {
	msg := strings.TrimSpace(stderr)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if msg == "" {
		return err
	}
	return fmt.Errorf("%w: %s", err, msg)
}
