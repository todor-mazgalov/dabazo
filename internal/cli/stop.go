// Package cli implements the stop command for dabazo.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"dabazo/internal/pkgmgr"
	"dabazo/internal/registry"
)

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop a database instance",
		Long:  "Stop the database service for a registered instance. Idempotent: stopping an already-stopped instance prints \"already stopped\" and exits 0.",
		Example: `  dabazo stop
  dabazo stop --name dev`,
		RunE: runStop,
	}
}

func runStop(cmd *cobra.Command, args []string) error {
	inst, err := registry.Resolve(flagName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitNotFound)
	}

	if inst.PackageManager == "external" {
		fmt.Fprintln(os.Stderr, "error: instance was added via `registry add`; dabazo does not manage its lifecycle")
		os.Exit(ExitUsage)
	}

	if _, err := pkgmgr.ByName(inst.PackageManager); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	eng, err := resolveEngine(inst.Engine)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	runner := newRunner()
	if err := eng.Stop(*inst, runner); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitPkgManager)
	}

	return nil
}
