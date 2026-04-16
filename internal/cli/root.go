// Package cli defines all command definitions and the entry point for the dabazo CLI.
package cli

import (
	"fmt"
	"os"
	"runtime/debug"
)

// version returns the module version embedded by Go at build time.
// When installed via `go install ...@v0.1.1` it returns "v0.1.1".
// For local builds without a tag it returns "(devel)".
func version() string {
	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "(devel)"
}

var (
	flagName   string
	flagEngine string
	flagPort   int
	flagYes    bool
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
			newCreateCommand(),
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
	for _, a := range os.Args[1:] {
		if a == "--version" || a == "-v" {
			fmt.Println("dabazo " + version())
			return
		}
		if a == "--" {
			break
		}
	}
	root := newRootCommand()
	if err := dispatch(root, os.Args); err != nil {
		os.Exit(1)
	}
}
