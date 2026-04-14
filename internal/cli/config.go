// Package cli implements the config command group for dabazo.
package cli

import (
	"fmt"
	"os"

	"dabazo/internal/registry"
	"dabazo/internal/secret"
)

// newConfigCommand creates the config command group with its subcommands.
func newConfigCommand() *command {
	return &command{
		name:  "config",
		use:   "config",
		short: "Configuration subcommands",
		subcommands: []*command{
			newConfigUserCommand(),
		},
	}
}

// newConfigUserCommand creates the "config user" subcommand descriptor.
func newConfigUserCommand() *command {
	return &command{
		name:  "user",
		use:   "user <username>",
		short: "Create a database role and credentials file",
		long: `Create a database role named <username> with a randomly generated password
and a database of the same name. Writes credentials to a file named <username>
in the current directory (mode 0600).`,
		example: `  dabazo config user alice
  dabazo config user bob --name dev`,
		run: runConfigUser,
	}
}

func runConfigUser(args []string) error {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "error: config user requires exactly 1 argument, got %d\n", len(args))
		os.Exit(ExitUsage)
	}
	username := args[0]

	inst, err := registry.Resolve(flagName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitNotFound)
	}

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

	runner := newRunner()
	if err := eng.CreateUser(*inst, username, password, runner); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitDBOperation)
	}

	content := fmt.Sprintf("DB_URL=jdbc:postgresql://%s:%d/%s\nDB_USER=%s\nDB_PASSWORD=%s\n",
		inst.Host, inst.Port, username, username, password)
	if err := os.WriteFile(username, []byte(content), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "error: writing credential file: %v\n", err)
		os.Exit(ExitGeneric)
	}

	fmt.Printf("Created role %q with database %q. Credentials written to ./%s\n", username, username, username)
	return nil
}
