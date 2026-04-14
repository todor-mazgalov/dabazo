// Package cli implements the migrate command for dabazo.
package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"dabazo/internal/registry"
)

var flagUser string

func newMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate <filepath>",
		Short: "Apply a SQL migration file to an instance",
		Long: `Apply the SQL file at <filepath> to the instance's database.
Requires --user to identify which role and credential file to use.`,
		Example: `  dabazo migrate ./V1__setup.sql --user alice
  dabazo migrate ./V2__data.sql --user bob --name dev`,
		Args: cobra.ExactArgs(1),
		RunE: runMigrate,
	}
	cmd.Flags().StringVar(&flagUser, "user", "", "database role to use (required)")
	return cmd
}

func runMigrate(cmd *cobra.Command, args []string) error {
	sqlFile := args[0]

	if flagUser == "" {
		fmt.Fprintln(os.Stderr, "error: --user is required for migrate")
		os.Exit(ExitUsage)
	}

	inst, err := registry.Resolve(flagName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitNotFound)
	}

	password, err := loadPassword(flagUser)
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
	if err := eng.ApplySQL(*inst, flagUser, password, sqlFile, runner); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitDBOperation)
	}

	fmt.Printf("Migration %q applied successfully.\n", sqlFile)
	return nil
}

// loadPassword reads the DB_PASSWORD from the credential file in the current directory.
func loadPassword(username string) (string, error) {
	f, err := os.Open(username)
	if err != nil {
		return "", fmt.Errorf("opening credential file %q: %w", username, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "DB_PASSWORD=") {
			return strings.TrimPrefix(line, "DB_PASSWORD="), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading credential file: %w", err)
	}
	return "", fmt.Errorf("DB_PASSWORD not found in credential file %q", username)
}
