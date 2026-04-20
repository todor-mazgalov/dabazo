// Package cli implements the migrate command for dabazo.
package cli

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/todor-mazgalov/dabazo/internal/credfmt"
	"github.com/todor-mazgalov/dabazo/internal/prompt"
	"github.com/todor-mazgalov/dabazo/internal/registry"
)

var (
	flagUser     string
	flagDatabase string
	flagSchema   string
)

// newMigrateCommand creates the migrate command descriptor.
func newMigrateCommand() *command {
	return &command{
		name:  "migrate",
		use:   "migrate <filepath>",
		short: "Apply a SQL migration file to an instance",
		long: `Apply the SQL file at <filepath> to the instance's database.
Requires --user to identify which role and credential file to use.
--database defaults to --user when omitted. --schema, if provided, is
applied as PostgreSQL search_path for the session. --host overrides the
instance host from the registry.`,
		example: `  dabazo migrate ./V1__setup.sql --user alice
  dabazo migrate ./V2__data.sql -u bob -n dev -db app -s public
  dabazo migrate ./V3.sql -u alice -h 10.0.0.5 -p 5433`,
		run: runMigrate,
		requiredFlags: []requiredFlag{
			{
				name:        "user",
				description: "Database role",
				isMissing:   func() bool { return flagUser == "" },
				set:         stringFlagSetter(&flagUser),
			},
		},
		localFlags: func(fs *flag.FlagSet) {
			fs.StringVar(&flagUser, "user", "", "database role to use (required)")
			fs.StringVar(&flagUser, "u", "", "short for --user")
			fs.StringVar(&flagDatabase, "database", "", "database name (defaults to --user)")
			fs.StringVar(&flagDatabase, "db", "", "short for --database")
			fs.StringVar(&flagSchema, "schema", "", "schema for session search_path")
			fs.StringVar(&flagSchema, "s", "", "short for --schema")
			fs.StringVar(&flagHost, "host", "", "override instance host")
			fs.StringVar(&flagHost, "h", "", "short for --host")
		},
	}
}

func runMigrate(args []string) error {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "error: migrate requires exactly 1 argument, got %d\n", len(args))
		os.Exit(ExitUsage)
	}
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
	printInstanceName(inst.Name)

	if flagHost != "" {
		inst.Host = flagHost
	}
	if flagPort != 0 {
		inst.Port = flagPort
	}

	database := flagDatabase
	if database == "" {
		database = flagUser
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
	if err := eng.ApplySQL(*inst, flagUser, password, database, flagSchema, sqlFile, runner); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitDBOperation)
	}

	fmt.Printf("Migration %q applied successfully.\n", sqlFile)
	return nil
}

// loadPassword returns the DB_PASSWORD for username. It first tries a credential
// file named <username> in the current directory; if that file does not exist it
// prompts the user for the password interactively with hidden input.
func loadPassword(username string) (string, error) {
	f, err := os.Open(username)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(os.Stderr, "No credential file %q found in current directory.\n", username)
		return prompt.ReadPassword(fmt.Sprintf("Password for %q: ", username), os.Stderr)
	}
	if err != nil {
		return "", fmt.Errorf("opening credential file %q: %w", username, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if v, ok := credfmt.ExtractValue(scanner.Text(), "DB_PASSWORD"); ok {
			return v, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading credential file: %w", err)
	}
	return "", fmt.Errorf("DB_PASSWORD not found in credential file %q", username)
}
