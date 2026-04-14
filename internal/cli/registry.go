// Package cli implements the registry command group for dabazo.
package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"dabazo/internal/engines"
	"dabazo/internal/registry"
)

var flagHost string

func newRegistryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Manage the instance registry",
	}
	cmd.AddCommand(newRegistryAddCmd())
	cmd.AddCommand(newRegistryRemoveCmd())
	return cmd
}

func newRegistryAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Register an externally-managed database instance",
		Long: `Register an already-existing database instance that was not installed by dabazo.
No packages are installed, no service is touched. The instance is recorded with
packageManager "external" and cannot be started/stopped/uninstalled by dabazo.`,
		Example: `  dabazo registry add --db postgres:16 --port 5432 --name legacy
  dabazo registry add --db postgres:16 --port 5432 --name remote --host 10.0.0.5`,
		RunE: runRegistryAdd,
	}
	cmd.Flags().StringVar(&flagHost, "host", "localhost", "host address for the instance")
	return cmd
}

func runRegistryAdd(cmd *cobra.Command, args []string) error {
	if flagDB == "" {
		fmt.Fprintln(os.Stderr, "error: --db is required for registry add")
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

	engineName, version := parseDB(flagDB)
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

func newRegistryRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove",
		Short: "Remove an instance from the registry",
		Long: `Remove an entry from the registry without uninstalling anything or touching
the database itself. Safe for both dabazo-installed and externally-added entries.
--name is always required.`,
		Example: `  dabazo registry remove --name legacy`,
		RunE:    runRegistryRemove,
	}
}

func runRegistryRemove(cmd *cobra.Command, args []string) error {
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
