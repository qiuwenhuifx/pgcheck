package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/xiongcccc/pgcheck/internal/app"
)

//go:embed SQL/*.sql
var sqlFS embed.FS

var (
	version = "2.0.0"
	commit  = ""
	date    = ""
)

func main() {
	cli := app.New(sqlFS, app.BuildInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	})
	if err := cli.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "pgcheck:", err)
		os.Exit(1)
	}
}
