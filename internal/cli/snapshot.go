// Package cli implements the snapshot command for dabazo.
package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/todor-mazgalov/dabazo/internal/prompt"
	"github.com/todor-mazgalov/dabazo/internal/registry"
)

var flagForce bool

// newSnapshotCommand creates the snapshot command descriptor.
func newSnapshotCommand() *command {
	return &command{
		name:  "snapshot",
		use:   "snapshot <db> <path>",
		short: "Dump a database to a plain SQL file",
		long: `Dump the entire database <db> to <path> as plain SQL (schema + data).
Prompts interactively for database user and password. The dump is importable
on another instance.`,
		example: `  dabazo snapshot alice /tmp/alice.sql
  dabazo snapshot mydb ./backup.sql --name dev --force`,
		run: runSnapshot,
		localFlags: func(fs *flag.FlagSet) {
			fs.BoolVar(&flagForce, "force", false, "overwrite existing output file")
		},
	}
}

func runSnapshot(args []string) error {
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "error: snapshot requires exactly 2 arguments, got %d\n", len(args))
		os.Exit(ExitUsage)
	}
	dbName := args[0]
	outPath := args[1]

	inst, err := registry.Resolve(flagName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitNotFound)
	}
	printInstanceName(inst.Name)

	if !flagForce {
		if _, err := os.Stat(outPath); err == nil {
			fmt.Fprintf(os.Stderr, "error: file %q already exists; use --force to overwrite\n", outPath)
			os.Exit(ExitGeneric)
		}
	}

	user, err := prompt.ReadLine("Database user: ", os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	password, err := prompt.ReadPassword("Password: ", os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	eng, err := resolveEngine(inst.Engine)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	runner := newRunner()
	if err := eng.Dump(*inst, dbName, user, password, outPath, runner); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitDBOperation)
	}

	fmt.Printf("Snapshot of %q written to %s\n", dbName, outPath)
	return nil
}
