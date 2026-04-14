// Package cli defines all cobra command definitions for the dabazo CLI.
package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	flagName string
	flagDB   string
	flagPort int
	flagYes  bool
)

// NewRootCmd creates the root cobra command with all subcommands registered.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "dabazo",
		Short: "Cross-platform CLI for installing, running, and operating database engines",
		Long: `dabazo is a cross-platform CLI tool for installing, running, and operating
database engines during development and debugging. MVP supports PostgreSQL.

Use the same CLI shape on Linux, macOS, and Windows to spin up a local database
instance, create users, apply migrations, and snapshot data for debugging.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&flagName, "name", "", "logical instance name")
	root.PersistentFlags().StringVar(&flagDB, "db", "", "engine[:version] (e.g. postgres:16)")
	root.PersistentFlags().IntVar(&flagPort, "port", 0, "TCP port the instance listens on")
	root.PersistentFlags().BoolVarP(&flagYes, "yes", "y", false, "skip confirmation prompts")

	root.AddCommand(
		newInstallCmd(),
		newStartCmd(),
		newStopCmd(),
		newConfigCmd(),
		newMigrateCmd(),
		newSnapshotCmd(),
		newRegistryCmd(),
		newListCmd(),
		newUninstallCmd(),
	)

	return root
}

// Execute runs the root command and exits with the appropriate code.
func Execute() {
	root := NewRootCmd()
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
