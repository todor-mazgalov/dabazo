// Package cli implements the snapshot command for dabazo.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"dabazo/internal/prompt"
	"dabazo/internal/registry"
)

var flagForce bool

func newSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot <db> <path>",
		Short: "Dump a database to a plain SQL file",
		Long: `Dump the entire database <db> to <path> as plain SQL (schema + data).
Prompts interactively for database user and password. The dump is importable
on another instance.`,
		Example: `  dabazo snapshot alice /tmp/alice.sql
  dabazo snapshot mydb ./backup.sql --name dev --force`,
		Args: cobra.ExactArgs(2),
		RunE: runSnapshot,
	}
	cmd.Flags().BoolVar(&flagForce, "force", false, "overwrite existing output file")
	return cmd
}

func runSnapshot(cmd *cobra.Command, args []string) error {
	dbName := args[0]
	outPath := args[1]

	inst, err := registry.Resolve(flagName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitNotFound)
	}

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
