// Package cli implements the delete command group for dabazo.
package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/todor-mazgalov/dabazo/internal/registry"
)

// newDeleteCommand creates the delete command group with its subcommands.
func newDeleteCommand() *command {
	return &command{
		name:  "delete",
		use:   "delete",
		short: "Delete users, databases, and schemas",
		subcommands: []*command{
			newDeleteUserCommand(),
			newDeleteDatabaseCommand(),
			newDeleteSchemaCommand(),
		},
	}
}

// newDeleteUserCommand creates the "delete user" subcommand descriptor.
func newDeleteUserCommand() *command {
	return &command{
		name:  "user",
		use:   "user <username>",
		short: "Drop a database role",
		long:  `Drop the database role named <username>. Prompts for confirmation before executing.`,
		example: `  dabazo delete user alice
  dabazo delete user bob --name dev -y`,
		run: runDeleteUser,
	}
}

func runDeleteUser(args []string) error {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "error: delete user requires exactly 1 argument, got %d\n", len(args))
		os.Exit(ExitUsage)
	}
	username := args[0]

	inst, err := registry.Resolve(flagName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitNotFound)
	}
	printInstanceName(inst.Name)

	eng, err := resolveEngine(inst.Engine)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	fmt.Printf("Will drop role %q from instance %q\n", username, inst.Name)
	confirmOrAbort(flagYes)

	runner := newRunner()
	if err := eng.DropUser(*inst, username, runner); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitDBOperation)
	}

	fmt.Printf("Dropped role %q.\n", username)
	return nil
}

// newDeleteDatabaseCommand creates the "delete database" subcommand descriptor.
func newDeleteDatabaseCommand() *command {
	return &command{
		name:  "database",
		use:   "database <database-name>",
		short: "Drop a database",
		long:  `Drop the database named <database-name>. Prompts for confirmation before executing.`,
		example: `  dabazo delete database app
  dabazo delete database reports --name dev -y`,
		run: runDeleteDatabase,
	}
}

func runDeleteDatabase(args []string) error {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "error: delete database requires exactly 1 argument, got %d\n", len(args))
		os.Exit(ExitUsage)
	}
	database := args[0]

	inst, err := registry.Resolve(flagName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitNotFound)
	}
	printInstanceName(inst.Name)

	eng, err := resolveEngine(inst.Engine)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	fmt.Printf("Will drop database %q from instance %q\n", database, inst.Name)
	confirmOrAbort(flagYes)

	runner := newRunner()
	if err := eng.DropDatabase(*inst, database, runner); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitDBOperation)
	}

	fmt.Printf("Dropped database %q.\n", database)
	return nil
}

// newDeleteSchemaCommand creates the "delete schema" subcommand descriptor.
func newDeleteSchemaCommand() *command {
	return &command{
		name:  "schema",
		use:   "schema <schema-name>",
		short: "Drop a schema from a database",
		long: `Drop the schema named <schema-name> from the database identified by --database,
connecting as the role identified by --user. Reads the user's password from a
credential file named after --user in the current directory.`,
		example: `  dabazo delete schema audit -db app -u alice
  dabazo delete schema public -db app -u alice --name dev -y`,
		run: runDeleteSchema,
		localFlags: func(fs *flag.FlagSet) {
			fs.StringVar(&flagUser, "user", "", "role to connect as (required)")
			fs.StringVar(&flagUser, "u", "", "short for --user")
			fs.StringVar(&flagDatabase, "database", "", "database containing the schema (required)")
			fs.StringVar(&flagDatabase, "db", "", "short for --database")
		},
		requiredFlags: []requiredFlag{
			{
				name:        "user",
				description: "Role to connect as",
				isMissing:   func() bool { return flagUser == "" },
				set:         stringFlagSetter(&flagUser),
			},
			{
				name:        "database",
				description: "Database name",
				isMissing:   func() bool { return flagDatabase == "" },
				set:         stringFlagSetter(&flagDatabase),
			},
		},
	}
}

func runDeleteSchema(args []string) error {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "error: delete schema requires exactly 1 argument, got %d\n", len(args))
		os.Exit(ExitUsage)
	}
	schema := args[0]

	if flagUser == "" {
		fmt.Fprintln(os.Stderr, "error: --user is required for delete schema")
		os.Exit(ExitUsage)
	}
	if flagDatabase == "" {
		fmt.Fprintln(os.Stderr, "error: --database is required for delete schema")
		os.Exit(ExitUsage)
	}

	inst, err := registry.Resolve(flagName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitNotFound)
	}
	printInstanceName(inst.Name)

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

	fmt.Printf("Will drop schema %q from database %q on instance %q\n", schema, flagDatabase, inst.Name)
	confirmOrAbort(flagYes)

	runner := newRunner()
	if err := eng.DropSchema(*inst, flagDatabase, schema, flagUser, password, runner); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitDBOperation)
	}

	fmt.Printf("Dropped schema %q from database %q.\n", schema, flagDatabase)
	return nil
}
