// Package cli implements the create command group for dabazo.
package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/todor-mazgalov/dabazo/internal/credfmt"
	"github.com/todor-mazgalov/dabazo/internal/registry"
	"github.com/todor-mazgalov/dabazo/internal/secret"
)

var (
	flagOutput    string
	flagURLFormat string
)

// newCreateCommand creates the create command group with its subcommands.
func newCreateCommand() *command {
	return &command{
		name:  "create",
		use:   "create",
		short: "Create users, databases, and schemas",
		subcommands: []*command{
			newCreateUserCommand(),
			newCreateDatabaseCommand(),
			newCreateSchemaCommand(),
		},
	}
}

// newCreateUserCommand creates the "create user" subcommand descriptor.
func newCreateUserCommand() *command {
	return &command{
		name:  "user",
		use:   "user <username>",
		short: "Create a database role and credentials file",
		long: `Create a database role named <username> with a randomly generated password
and a database of the same name. Writes credentials to a file named <username>
in the current directory (mode 0600).`,
		example: `  dabazo create user alice
  dabazo create user bob --name dev
  dabazo create user alice --output bash
  dabazo create user alice -o pwsh`,
		run: runCreateUser,
		localFlags: func(fs *flag.FlagSet) {
			fs.StringVar(&flagOutput, "output", "java", "credential file format (java, shell, bash, bat, cmd, pwsh, powershell)")
			fs.StringVar(&flagOutput, "o", "java", "short for --output")
			fs.StringVar(&flagURLFormat, "url-format", "jdbc", "URL format for DB_URL (jdbc, plain)")
			fs.StringVar(&flagURLFormat, "uf", "jdbc", "short for --url-format")
		},
	}
}

func runCreateUser(args []string) error {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "error: create user requires exactly 1 argument, got %d\n", len(args))
		os.Exit(ExitUsage)
	}
	username := args[0]

	inst, err := registry.Resolve(flagName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitNotFound)
	}
	printInstanceName(inst.Name)

	if _, err := os.Stat(username); err == nil {
		fmt.Fprintf(os.Stderr, "error: credential file %q already exists; refusing to overwrite\n", username)
		os.Exit(ExitGeneric)
	}

	password, err := secret.GeneratePassword(32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	eng, err := resolveEngine(inst.Engine)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	outFmt, err := credfmt.Parse(flagOutput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitUsage)
	}

	urlFmt, err := credfmt.ParseURLFormat(flagURLFormat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitUsage)
	}

	runner := newRunner()
	if err := eng.CreateUser(*inst, username, password, runner); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitDBOperation)
	}

	dbURL := credfmt.FormatURL(urlFmt, inst.Engine, inst.Host, inst.Port, username)
	kvs := []credfmt.KV{
		{Key: "DB_URL", Value: dbURL},
		{Key: "DB_USER", Value: username},
		{Key: "DB_PASSWORD", Value: password},
	}
	content := credfmt.Render(outFmt, kvs)
	if err := os.WriteFile(username, []byte(content), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "error: writing credential file: %v\n", err)
		os.Exit(ExitGeneric)
	}

	fmt.Printf("Created role %q with database %q. Credentials written to ./%s\n", username, username, username)
	return nil
}

// newCreateDatabaseCommand creates the "create database" subcommand descriptor.
func newCreateDatabaseCommand() *command {
	return &command{
		name:  "database",
		use:   "database <database-name>",
		short: "Create a database owned by an existing role",
		long: `Create a database owned by the role identified by --user. The role must
already exist (created by 'dabazo create user' or externally). Runs as the
PostgreSQL superuser and does not require a credential file.`,
		example: `  dabazo create database app -u alice
  dabazo create database reports -u alice --name dev`,
		run: runCreateDatabase,
		localFlags: func(fs *flag.FlagSet) {
			fs.StringVar(&flagUser, "user", "", "owner role for the new database (required)")
			fs.StringVar(&flagUser, "u", "", "short for --user")
		},
		requiredFlags: []requiredFlag{
			{
				name:        "user",
				description: "Owner role",
				isMissing:   func() bool { return flagUser == "" },
				set:         stringFlagSetter(&flagUser),
			},
		},
	}
}

func runCreateDatabase(args []string) error {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "error: create database requires exactly 1 argument, got %d\n", len(args))
		os.Exit(ExitUsage)
	}
	database := args[0]

	if flagUser == "" {
		fmt.Fprintln(os.Stderr, "error: --user is required for create database")
		os.Exit(ExitUsage)
	}

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

	runner := newRunner()
	if err := eng.CreateDatabase(*inst, database, flagUser, runner); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitDBOperation)
	}

	fmt.Printf("Created database %q owned by %q.\n", database, flagUser)
	return nil
}

// newCreateSchemaCommand creates the "create schema" subcommand descriptor.
func newCreateSchemaCommand() *command {
	return &command{
		name:  "schema",
		use:   "schema <schema-name>",
		short: "Create a schema inside an existing database",
		long: `Create a schema inside the database identified by --database, connecting
as the role identified by --user. Reads the user's password from a credential
file named after --user in the current directory.`,
		example: `  dabazo create schema public -db app -u alice
  dabazo create schema audit -db app -u alice --name dev`,
		run: runCreateSchema,
		localFlags: func(fs *flag.FlagSet) {
			fs.StringVar(&flagUser, "user", "", "role to connect as (required)")
			fs.StringVar(&flagUser, "u", "", "short for --user")
			fs.StringVar(&flagDatabase, "database", "", "database to create the schema in (required)")
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

func runCreateSchema(args []string) error {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "error: create schema requires exactly 1 argument, got %d\n", len(args))
		os.Exit(ExitUsage)
	}
	schema := args[0]

	if flagUser == "" {
		fmt.Fprintln(os.Stderr, "error: --user is required for create schema")
		os.Exit(ExitUsage)
	}
	if flagDatabase == "" {
		fmt.Fprintln(os.Stderr, "error: --database is required for create schema")
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

	runner := newRunner()
	if err := eng.CreateSchema(*inst, flagDatabase, schema, flagUser, password, runner); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitDBOperation)
	}

	fmt.Printf("Created schema %q in database %q.\n", schema, flagDatabase)
	return nil
}
