// Package cli implements the uninstall command for dabazo.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"dabazo/internal/pkgmgr"
	"dabazo/internal/registry"
)

var flagPurge bool

func newUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall a database instance and remove it from the registry",
		Long: `Stop the instance (if running), uninstall the packages via the same package
manager (with confirmation), and remove the registry entry. Does not delete
the data directory unless --purge is passed.`,
		Example: `  dabazo uninstall --name dev
  dabazo uninstall --name dev --purge -y`,
		RunE: runUninstall,
	}
	cmd.Flags().BoolVar(&flagPurge, "purge", false, "also delete the data directory")
	return cmd
}

func runUninstall(cmd *cobra.Command, args []string) error {
	inst, err := registry.Resolve(flagName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitNotFound)
	}

	if inst.PackageManager == "external" {
		fmt.Fprintln(os.Stderr, "error: instance was added via `registry add`; dabazo does not manage its lifecycle")
		os.Exit(ExitUsage)
	}

	pm, err := pkgmgr.ByName(inst.PackageManager)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	eng, err := resolveEngine(inst.Engine)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	plan, err := eng.UninstallPlan(*inst, pm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	printPlan("uninstall", plan, inst.Name)
	confirmOrAbort(flagYes)

	runner := newRunner()

	// Stop if running.
	if eng.IsRunning(*inst) {
		if err := eng.Stop(*inst, runner); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to stop instance: %v\n", err)
		}
	}

	// Run uninstall commands.
	for _, c := range plan.Commands {
		if len(c) == 0 {
			continue
		}
		if _, err := runner.Run(c[0], c[1:]...); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(ExitPkgManager)
		}
	}

	// Purge data directory if requested.
	if flagPurge && inst.DataDir != "" {
		if err := os.RemoveAll(inst.DataDir); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to remove data directory: %v\n", err)
		} else {
			fmt.Printf("Data directory %s removed.\n", inst.DataDir)
		}
	}

	if err := registry.Remove(inst.Name); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	fmt.Printf("Instance %q uninstalled and removed from registry.\n", inst.Name)
	return nil
}
