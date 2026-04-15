// Package cli implements the registry command group for dabazo.
package cli

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/todor-mazgalov/dabazo/internal/engines"
	"github.com/todor-mazgalov/dabazo/internal/registry"
)

var flagHost string

// newRegistryCommand creates the registry command group with add and remove subcommands.
func newRegistryCommand() *command {
	return &command{
		name:  "registry",
		use:   "registry",
		short: "Manage the instance registry",
		subcommands: []*command{
			newRegistryAddCommand(),
			newRegistryRemoveCommand(),
		},
	}
}

// newRegistryAddCommand creates the "registry add" subcommand descriptor.
func newRegistryAddCommand() *command {
	return &command{
		name:  "add",
		use:   "add",
		short: "Register an externally-managed database instance",
		long: `Register an already-existing database instance that was not installed by dabazo.
No packages are installed, no service is touched. The instance is recorded with
packageManager "external" and cannot be started/stopped/uninstalled by dabazo.`,
		example: `  dabazo registry add --db postgres:16 --port 5432 --name legacy
  dabazo registry add --db postgres:16 --port 5432 --name remote --host 10.0.0.5`,
		run: runRegistryAdd,
		localFlags: func(fs *flag.FlagSet) {
			fs.StringVar(&flagHost, "host", "localhost", "host address for the instance")
		},
	}
}

func runRegistryAdd(args []string) error {
	if flagEngine == "" {
		fmt.Fprintln(os.Stderr, "error: --engine is required for registry add")
		os.Exit(ExitUsage)
	}
	if flagPort == 0 {
		fmt.Fprintln(os.Stderr, "error: --port is required for registry add")
		os.Exit(ExitUsage)
	}
	if flagName == "" {
		fmt.Fprintln(os.Stderr, "error: --name is required for registry add")
		os.Exit(ExitUsage)
	}

	engineName, version := parseDB(flagEngine)
	if version == "" {
		version = "unknown"
	}

	inst := engines.Instance{
		Name:           flagName,
		Engine:         engineName,
		Version:        version,
		Port:           flagPort,
		Host:           flagHost,
		InstalledAt:    time.Now().UTC().Format(time.RFC3339),
		PackageManager: "external",
	}

	if err := registry.Add(inst); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitAlreadyExists)
	}

	fmt.Printf("Instance %q registered as external.\n", flagName)
	return nil
}

// newRegistryRemoveCommand creates the "registry remove" subcommand descriptor.
func newRegistryRemoveCommand() *command {
	return &command{
		name:  "remove",
		use:   "remove",
		short: "Remove an instance from the registry",
		long: `Remove an entry from the registry without uninstalling anything or touching
the database itself. Safe for both dabazo-installed and externally-added entries.
--name is always required.`,
		example: `  dabazo registry remove --name legacy`,
		run:     runRegistryRemove,
	}
}

func runRegistryRemove(args []string) error {
	if flagName == "" {
		fmt.Fprintln(os.Stderr, "error: --name is required for registry remove")
		os.Exit(ExitUsage)
	}

	if err := registry.Remove(flagName); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitNotFound)
	}

	fmt.Printf("Instance %q removed from registry.\n", flagName)
	return nil
}
