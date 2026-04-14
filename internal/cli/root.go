// Package cli defines all command definitions and the entry point for the dabazo CLI.
package cli

import (
	"os"
)

var (
	flagName string
	flagDB   string
	flagPort int
	flagYes  bool
)

// newRootCommand builds the root command tree with all subcommands registered.
func newRootCommand() *command {
	return &command{
		name: "dabazo",
		use:  "dabazo",
		short: "Cross-platform CLI for installing, running, and operating database engines",
		long: `dabazo is a cross-platform CLI tool for installing, running, and operating
database engines during development and debugging. MVP supports PostgreSQL.

Use the same CLI shape on Linux, macOS, and Windows to spin up a local database
instance, create users, apply migrations, and snapshot data for debugging.`,
		subcommands: []*command{
			newInstallCommand(),
			newStartCommand(),
			newStopCommand(),
			newConfigCommand(),
			newMigrateCommand(),
			newSnapshotCommand(),
			newRegistryCommand(),
			newListCommand(),
			newUninstallCommand(),
		},
	}
}

// Execute runs the root command and exits with code 1 on unhandled errors.
func Execute() {
	root := newRootCommand()
	if err := dispatch(root, os.Args); err != nil {
		os.Exit(1)
	}
}
